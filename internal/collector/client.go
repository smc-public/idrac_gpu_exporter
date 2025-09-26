package collector

import (
	"fmt"
	"strings"

	"github.com/smc-public/idrac_gpu_exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	UNKNOWN = iota
	DELL
	HPE
	LENOVO
	INSPUR
	H3C
	INVENTEC
	FUJITSU
	SUPERMICRO
)

type Client struct {
	redfish     *Redfish
	vendor      int
	systemPath  string
	procPath    string
}

type GPUInfo struct {
	Id                    string
	Manufacturer          string
	Model                 string
	PartNumber            string
	SerialNumber          string
	UUID                 string
}

func NewClient(h *config.HostConfig) *Client {
	client := &Client{
		redfish: NewRedfish(
			h.Scheme,
			h.Hostname,
			h.Username,
			h.Password,
		),
	}

	client.redfish.CreateSession()
	ok := client.findAllEndpoints()
	if !ok {
		client.redfish.DeleteSession()
		return nil
	}

	return client
}

func (client *Client) findAllEndpoints() bool {
	var root V1Response
	var group GroupResponse
	var system SystemResponse
	var ok bool

	// Root
	ok = client.redfish.Get(redfishRootPath, &root)
	if !ok {
		return false
	}

	// System
	ok = client.redfish.Get(root.Systems.OdataId, &group)
	if !ok {
		return false
	}

	client.systemPath = group.Members[0].OdataId

	ok = client.redfish.Get(client.systemPath, &system)
	if !ok {
		return false
	}

	client.procPath = system.Processors.OdataId

	// Vendor
	m := strings.ToLower(system.Manufacturer)
	if strings.Contains(m, "dell") || strings.Contains(m, "sustainable"){
		client.vendor = DELL
	} else if strings.Contains(m, "hpe") {
		client.vendor = HPE
	} else if strings.Contains(m, "lenovo") {
		client.vendor = LENOVO
	} else if strings.Contains(m, "inspur") {
		client.vendor = INSPUR
	} else if strings.Contains(m, "h3c") {
		client.vendor = H3C
	} else if strings.Contains(m, "inventec") {
		client.vendor = INVENTEC
	} else if strings.Contains(m, "fujitsu") {
		client.vendor = FUJITSU
	} else if strings.Contains(m, "supermicro") {
		client.vendor = SUPERMICRO
	}

	return true
}

func (client *Client) RefreshGPUs(mc *Collector, ch chan<- prometheus.Metric) bool {
	group := GroupResponse{}
	ok := client.redfish.Get(client.procPath, &group)
	if !ok {
		return false
	}

	// Get inventory information for Dell GPUs

	dellVideo := DellVideo{}

	if client.vendor == DELL {
		// Get dell video inventory

		dellVideoPath := fmt.Sprintf("%s/Oem/Dell/DellVideo", client.systemPath)
		client.redfish.Get(dellVideoPath, &dellVideo)
		dellGPUSensorPath := fmt.Sprintf("%s/Oem/Dell/DellGPUSensors", client.systemPath)

		// Get dell GPU sensor metrics

		dellGPUSensors := DellGPUSensors{}
		if ok := client.redfish.Get(dellGPUSensorPath, &dellGPUSensors); ok {
			for _, v := range dellGPUSensors.Members {
				mc.NewBoardPowerSupplyStatus(ch, &v)
				mc.NewMemoryTemperatureCelsius(ch, &v)
				mc.NewPowerBrakeStatus(ch, &v)
				mc.NewPrimaryGPUTemperatureCelsius(ch, &v)
				mc.NewThermalAlertStatus(ch, &v)
			}
		}
	}

	// Get GPU metrics

	for _, c := range group.Members.GetLinks() {
		resp := GPU{}
		ok = client.redfish.Get(c, &resp)
		if !ok {
			continue
		}

		if resp.ProcessorType != "GPU" {
			continue
		}

		if resp.Status.State != StateEnabled {
			continue
		}

		gpuInfo := GPUInfo{}
		gpuInfo.Id = resp.Id
		gpuInfo.Manufacturer = resp.Manufacturer
		gpuInfo.Model = resp.Model
		gpuInfo.PartNumber = resp.PartNumber

		if client.vendor == DELL {
			for _, v := range dellVideo.Members {
				if v.Id == resp.Id {
					gpuInfo.UUID = v.GPUGUID
					gpuInfo.SerialNumber = v.SerialNumber
					mc.NewGPUState(ch, &v)
					mc.NewGPUHealth(ch, &v)
					break
				}
			}
		}

		mc.NewGPUInfo(ch, &gpuInfo)

		if resp.Metrics.OdataId != "" {
			gpuMetrics := GPUMetrics{}
			ok = client.redfish.Get(resp.Metrics.OdataId, &gpuMetrics)
			if !ok {
				break
			}

			mc.NewGPUBandwidthPercent(ch, &gpuMetrics)
			mc.NewGPUConsumedPowerWatt(ch, &gpuMetrics)
			mc.NewGPUOperatingSpeedMHz(ch, &gpuMetrics)

			// // NVIDIA
            // TODO: check this

			// if metrics.Oem.Nvidia != nil {
			// 	mc.NewGPUUtilization(ch, metrics.Oem.Nvidia.UtilizationPercentage, resp.Id)
			// 	mc.NewGPUTemperature(ch, metrics.Oem.Nvidia.TemperatureCelsius, resp.Id)
			// 	mc.NewGPUMemoryTotal(ch, float64(metrics.Oem.Nvidia.MemoryTotalMiB*1024*1024), resp.Id)
			// 	mc.NewGPUMemoryUsed(ch, float64(metrics.Oem.Nvidia.MemoryUsedMiB*1024*1024), resp.Id)
			// 	mc.NewGPUMemoryFree(ch, float64(metrics.Oem.Nvidia.MemoryFreeMiB*1024*1024), resp.Id)
			// 	mc.NewGPUMemoryUtilization(ch, metrics.Oem.Nvidia.MemoryUtilizationPercentage, resp.Id)
			// }
		}

		if resp.MemorySummary.Metrics.OdataId != "" {
			gpuMemoryMetrics := GPUMemoryMetrics{}
			ok = client.redfish.Get(resp.MemorySummary.Metrics.OdataId, &gpuMemoryMetrics)
			if !ok {
				break
			}

			mc.NewGPUMemoryBandwidthPercent(ch, resp.Id, &gpuMemoryMetrics)
			mc.NewGPUMemoryOperatingSpeedMHz(ch, resp.Id, &gpuMemoryMetrics)
		}
	}

	return true
}
