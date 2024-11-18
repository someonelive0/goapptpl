package main

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/jaypipes/ghw"
	log "github.com/sirupsen/logrus"
)

type HardwareHandler struct {
}

// r := app.Group("/hardware")
func (p *HardwareHandler) AddRouter(r fiber.Router) error {
	log.Info("MysqlHandler AddRouter")

	r.Get("", p.hdHandler)
	r.Get("/", p.hdHandler)
	r.Get("/cpu", p.cpuHandler)
	r.Get("/mem", p.memHandler)
	r.Get("/block", p.blockHandler)
	r.Get("/network", p.networkHandler)
	r.Get("/pci", p.pciHandler)
	r.Get("/gpu", p.gpuHandler)
	r.Get("/chassis", p.chassisHandler)
	r.Get("/bios", p.biosHandler)
	r.Get("/baseboard", p.baseboardHandler)
	r.Get("/product", p.productHandler)

	return nil
}

// GET /hardware/cpu
func (p *HardwareHandler) hdHandler(c fiber.Ctx) error {
	c.Context().SetContentType("text/html")
	c.WriteString(`<html><body><h1>Hardware</h1>
	<a href="/hardware/cpu">cpu</a><br>
	<a href="/hardware/mem">mem</a><br>
	<a href="/hardware/block">block</a><br>
	<a href="/hardware/network">network</a><br>
	<a href="/hardware/pci">pci</a><br>
	<a href="/hardware/gpu">gpu</a><br>
	<a href="/hardware/chassis">chassis</a><br>
	<a href="/hardware/bios">bios</a><br>
	<a href="/hardware/baseboard">baseboard</a><br>
	<a href="/hardware/product">product</a><br>
	</body></html>`)
	return nil
}

// GET /hardware/cpu
func (p *HardwareHandler) cpuHandler(c fiber.Ctx) error {
	cpu, err := ghw.CPU()
	if err != nil {
		log.Errorf("Error getting CPU info: %v", err)
		return err
	}
	// c.WriteString(fmt.Sprintf("cpu %#v\n", cpu))

	c.Context().SetContentType("application/json")
	b, _ := json.Marshal(cpu)
	c.Write(b)

	// for _, proc := range cpu.Processors {
	// 	c.WriteString(fmt.Sprintf("proc %v\n", proc))
	// 	for _, core := range proc.Cores {
	// 		c.WriteString(fmt.Sprintf("core  %v\n", core))
	// 	}
	// 	if len(proc.Capabilities) > 0 {
	// 		// pretty-print the (large) block of capability strings into rows
	// 		// of 6 capability strings
	// 		rows := int(math.Ceil(float64(len(proc.Capabilities)) / float64(6)))
	// 		for row := 1; row < rows; row = row + 1 {
	// 			rowStart := (row * 6) - 1
	// 			rowEnd := int(math.Min(float64(rowStart+6), float64(len(proc.Capabilities))))
	// 			rowElems := proc.Capabilities[rowStart:rowEnd]
	// 			capStr := strings.Join(rowElems, " ")
	// 			if row == 1 {
	// 				fmt.Printf("  capabilities: [%s\n", capStr)
	// 			} else if rowEnd < len(proc.Capabilities) {
	// 				fmt.Printf("                 %s\n", capStr)
	// 			} else {
	// 				fmt.Printf("                 %s]\n", capStr)
	// 			}
	// 		}
	// 	}
	// }

	return nil
}

// GET /hardware/mem
func (p *HardwareHandler) memHandler(c fiber.Ctx) error {
	memory, err := ghw.Memory()
	if err != nil {
		log.Errorf("Error getting memory info: %v", err)
		return err
	}

	c.WriteString(fmt.Sprintf(`{ "memo": "%s", `, memory.String()))

	c.Context().SetContentType("application/json")
	b, _ := json.Marshal(memory)
	c.Write(b[1:])

	return nil
}

// GET /hardware/block
func (p *HardwareHandler) blockHandler(c fiber.Ctx) error {
	block, err := ghw.Block()
	if err != nil {
		log.Errorf("Error getting block storage info: %v", err)
		return err
	}

	c.Context().SetContentType("application/json")
	b, _ := json.Marshal(block)
	c.Write(b)

	// for _, disk := range block.Disks {
	// 	fmt.Printf("2 %v\n", disk)
	// 	for _, part := range disk.Partitions {
	// 		fmt.Printf("3  %v\n", part)
	// 	}
	// }

	return nil
}

// GET /hardware/network
func (p *HardwareHandler) networkHandler(c fiber.Ctx) error {
	net, err := ghw.Network()
	if err != nil {
		log.Errorf("Error getting network info: %v", err)
		return err
	}

	c.Context().SetContentType("application/json")
	b, _ := json.Marshal(net)
	c.Write(b)

	// for _, nic := range net.NICs {
	// 	fmt.Printf("2  %v\n", nic)

	// 	enabledCaps := make([]int, 0)
	// 	for x, cap := range nic.Capabilities {
	// 		if cap.IsEnabled {
	// 			enabledCaps = append(enabledCaps, x)
	// 		}
	// 	}
	// 	if len(enabledCaps) > 0 {
	// 		fmt.Printf("  enabled capabilities:\n")
	// 		for _, x := range enabledCaps {
	// 			fmt.Printf("   - %s\n", nic.Capabilities[x].Name)
	// 		}
	// 	}
	// }

	return nil
}

// GET /hardware/pci
func (p *HardwareHandler) pciHandler(c fiber.Ctx) error {
	pci, err := ghw.PCI()
	if err != nil {
		log.Errorf("Error getting PCI info: %v", err)
		return err
	}
	fmt.Printf("host PCI devices:\n")
	fmt.Println("====================================================")

	c.Context().SetContentType("application/json")
	b, _ := json.Marshal(pci)
	c.Write(b)

	// for _, device := range pci.Devices {
	// 	vendor := device.Vendor
	// 	vendorName := vendor.Name
	// 	if len(vendor.Name) > 20 {
	// 		vendorName = string([]byte(vendorName)[0:17]) + "..."
	// 	}
	// 	product := device.Product
	// 	productName := product.Name
	// 	if len(product.Name) > 40 {
	// 		productName = string([]byte(productName)[0:37]) + "..."
	// 	}
	// 	fmt.Printf("%-12s\t%-20s\t%-40s\n", device.Address, vendorName, productName)
	// }

	return nil
}

// GET /hardware/gpu
func (p *HardwareHandler) gpuHandler(c fiber.Ctx) error {
	gpu, err := ghw.GPU()
	if err != nil {
		log.Errorf("Error getting GPU info: %v", err)
		return err
	}

	c.Context().SetContentType("application/json")
	b, _ := json.Marshal(gpu)
	c.Write(b)

	// for _, card := range gpu.GraphicsCards {
	// 	fmt.Printf(" %v\n", card)
	// }

	return nil
}

// GET /hardware/chassis
func (p *HardwareHandler) chassisHandler(c fiber.Ctx) error {
	chassis, err := ghw.Chassis()
	if err != nil {
		log.Errorf("Error getting chassis info: %v", err)
		return err
	}

	c.Context().SetContentType("application/json")
	b, _ := json.Marshal(chassis)
	c.Write(b)

	return nil
}

// GET /hardware/bios
func (p *HardwareHandler) biosHandler(c fiber.Ctx) error {
	bios, err := ghw.BIOS()
	if err != nil {
		fmt.Printf("Error getting BIOS info: %v", err)
	}

	c.Context().SetContentType("application/json")
	b, _ := json.Marshal(bios)
	c.Write(b)

	return nil
}

// GET /hardware/baseboard
func (p *HardwareHandler) baseboardHandler(c fiber.Ctx) error {
	baseboard, err := ghw.Baseboard()
	if err != nil {
		fmt.Printf("Error getting baseboard info: %v", err)
	}

	c.Context().SetContentType("application/json")
	b, _ := json.Marshal(baseboard)
	c.Write(b)

	return nil
}

// GET /hardware/product
func (p *HardwareHandler) productHandler(c fiber.Ctx) error {
	product, err := ghw.Product()
	if err != nil {
		fmt.Printf("Error getting product info: %v", err)
	}

	c.Context().SetContentType("application/json")
	b, _ := json.Marshal(product)
	c.Write(b)

	return nil
}
