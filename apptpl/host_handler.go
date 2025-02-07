package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/patrickmn/go-cache"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	psnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	log "github.com/sirupsen/logrus"
)

type HostHandler struct {
	Mycache *cache.Cache
}

// r := app.Group("/host")
func (p *HostHandler) AddRouter(r fiber.Router) error {
	log.Info("HostHandler AddRouter")

	r.Get("", p.homeHandler)
	r.Get("/", p.homeHandler)
	r.Get("/os", p.osHandler)
	r.Get("/user", p.userHandler)
	r.Get("/cpu", p.cpuHandler)
	r.Get("/loading", p.loadingHandler)
	r.Get("/mem", p.memHandler)
	r.Get("/disk", p.diskHandler)
	r.Get("/net", p.netHandler)
	r.Get("/process", p.processHandler)

	return nil
}

// GET /host
func (p *HostHandler) homeHandler(c fiber.Ctx) error {
	c.Response().Header.Set("Content-Type", "text/html")
	c.WriteString(`<html><body><h1>Host Information</h1>
	<a href="/host/os">os</a><br>
	<a href="/host/user">user</a><br>
	<a href="/host/cpu">cpu</a><br>
	<a href="/host/loading">loading</a><br>
	<a href="/host/mem">mem</a><br>
	<a href="/host/disk">disk</a><br>
	<a href="/host/net">net</a><br>
	<a href="/host/process">process</a><br>
	</body></html>`)
	return nil
}

// GET /host/os
func (p *HostHandler) osHandler(c fiber.Ctx) error {
	boot_timestamp, _ := host.BootTime()
	t := time.Unix(int64(boot_timestamp), 0)
	s := fmt.Sprintf(`{"boot_time": "%s", `, t.Local().Format(time.RFC3339))

	// uptime, _ := host.Uptime()
	// s += fmt.Sprintf(`"uptime": "%v", `, uptime)

	// hostid, _ := host.HostID()
	// s += fmt.Sprintf(`"hostid": "%s", `, hostid)
	// arch, _ := host.KernelArch()
	// s += fmt.Sprintf(`"arch": "%s", `, arch)
	// kernel_version, _ := host.KernelVersion()
	// s += fmt.Sprintf(`"kernel_version": "%s", `, kernel_version)

	// platform, family, version, _ := host.PlatformInformation()
	// s += fmt.Sprintf(`"platform": "%s", `, platform)
	// s += fmt.Sprintf(`"family": "%s", `, family)
	// s += fmt.Sprintf(`"version": "%s"}`, version)

	c.Response().Header.Set("Content-Type", "application/json")
	c.WriteString(s)

	info, _ := host.Info()
	b, _ := json.MarshalIndent(info, "", " ")
	c.Write(b[1:])

	return nil
}

// GET /host/user
func (p *HostHandler) userHandler(c fiber.Ctx) error {
	users, err := host.Users()
	if err != nil {
		log.Warnf("get host users failed: %v", err)
		return err
	}

	c.Response().Header.Set("Content-Type", "application/json")
	b, _ := json.MarshalIndent(users, "", " ")
	c.Write(b)

	return nil
}

// GET /host/cpu
func (p *HostHandler) cpuHandler(c fiber.Ctx) error {
	physicalCnt, _ := cpu.Counts(false)
	logicalCnt, _ := cpu.Counts(true)
	s := fmt.Sprintf(`{"physical": %d, "logical": %d, `, physicalCnt, logicalCnt)

	// totalPercent, _ := cpu.Percent(3*time.Second, false)
	// perPercents, _ := cpu.Percent(3*time.Second, true)
	// fmt.Printf("total percent:%v per percents:%v", totalPercent, perPercents)

	infos, _ := cpu.Info()
	s += `"info": [`
	for i, info := range infos {
		if i > 0 {
			s += ", "
		}
		b, _ := json.MarshalIndent(info, "", " ")
		s += string(b)
	}
	s += `]}`

	// timescpu, _ := cpu.Times(true)
	// for _, info := range timescpu {
	// 	data, _ := json.MarshalIndent(info, "", " ")
	// 	fmt.Print(string(data))
	// }

	c.Response().Header.Set("Content-Type", "application/json")
	c.WriteString(s)

	return nil
}

// GET /host/loading
func (p *HostHandler) loadingHandler(c fiber.Ctx) error {
	s := fmt.Sprintf(`{"timestamp": "%s", `, time.Now().Format(time.RFC3339))
	s += `"measurement": "MiB", `
	v, _ := mem.VirtualMemory()
	s += fmt.Sprintf(`"mem_total": %v, `, v.Total)
	s += fmt.Sprintf(`"mem_available": %v, `, v.Available)
	s += fmt.Sprintf(`"mem_used_percent": %.2f, `, v.UsedPercent)

	totalPercent, _ := cpu.Percent(3*time.Second, false)
	perPercents, _ := cpu.Percent(3*time.Second, true)
	// fmt.Printf("total percent:%v per percents:%v", totalPercent, perPercents)
	s += `"cpu_percent": [ `
	for i := range totalPercent {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%.2f", totalPercent[i])
	}

	s += `] ,"cpu_per_percent": [ `
	for i := range perPercents {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%.2f", perPercents[i])
	}

	s += `]}`

	c.Response().Header.Set("Content-Type", "application/json")
	c.WriteString(s)

	return nil
}

// GET /host/mem
func (p *HostHandler) memHandler(c fiber.Ctx) error {
	v, _ := mem.VirtualMemory()
	//fmt.Printf("Total: %v, Available: %v, UsedPercent:%f%%\n", v.Total, v.Available, v.UsedPercent)
	s := fmt.Sprintf("%v", v)

	c.Response().Header.Set("Content-Type", "application/json")
	c.WriteString(s)

	return nil
}

// GET /host/disk
func (p *HostHandler) diskHandler(c fiber.Ctx) error {
	partitions, _ := disk.Partitions(true)
	b, _ := json.MarshalIndent(partitions, "", " ")
	c.Response().Header.Set("Content-Type", "application/json")
	c.Write(b)

	for _, partition := range partitions {
		diskInfo, _ := disk.Usage(partition.Mountpoint)
		b, _ = json.MarshalIndent(diskInfo, "", " ")
		c.Write(b)
	}

	return nil
}

// GET /host/net
func (p *HostHandler) netHandler(c fiber.Ctx) error {
	connections, _ := psnet.Connections("all")

	c.Response().Header.Set("Content-Type", "application/json")
	b, _ := json.Marshal(connections)
	c.Write(b)

	conntraks, err := psnet.ConntrackStats(false)
	if err != nil {
		log.Warnf("get ConntrackStats failed: %v", err)
	} else {
		b, _ = json.Marshal(conntraks)
		c.Write(b)
	}

	filterCounters, err := psnet.FilterCounters()
	if err != nil {
		log.Warnf("get FilterCounters failed: %v", err)
	} else {
		b, _ = json.Marshal(filterCounters)
		c.Write(b)
	}

	ioCounters, err := psnet.IOCounters(false)
	if err != nil {
		log.Warnf("get IOCounters failed: %v", err)
	} else {
		b, _ = json.Marshal(ioCounters)
		c.Write(b)
	}

	return nil
}

// GET /host/process
func (p *HostHandler) processHandler(c fiber.Ctx) error {
	pids, _ := process.Pids()

	c.Response().Header.Set("Content-Type", "application/json")
	b, _ := json.Marshal(pids)
	c.Write(b)

	return nil
}
