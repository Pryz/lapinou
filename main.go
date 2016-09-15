package main

import (
	"flag"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Pryz/libvirt-go"
	log "github.com/Sirupsen/logrus"
)

// CPU struct is used to create the CPU topology
type CPU struct {
	Name        string
	Id          int
	ThreadsList string
}

// Sorting CPUs by ID
type ById []CPU

func (a ById) Len() int           { return len(a) }
func (a ById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ById) Less(i, j int) bool { return a[i].Id < a[j].Id }

// Get a slice of CPUs defining the CPU topology with Threads Siblings list
//
// Example :
// [ {cpu0 0 0,24 }
// 	 {cpu1 1 1,25 }
// 	 {cpu2 2 2,26 }
// 	 {cpu3 3 3,27 }
// 	 {cpu4 4 4,28 }
// 	 {cpu5 5 5,29 }
// 	 {cpu6 6 6,30 }
// 	 {cpu7 7 7,31 }
// 	 {cpu8 8 8,32 }
// 	 {cpu9 9 9,33 } ]
//
func get_cpu_topo() (error, []CPU) {
	var cpus []CPU
	CPU_PATH := "/sys/devices/system/cpu"
	re, _ := regexp.Compile("^cpu[0-9]+")
	cpu_ds, _ := ioutil.ReadDir(CPU_PATH)

	for _, cpu_d := range cpu_ds {
		if re.MatchString(cpu_d.Name()) {
			tsiblings, err := ioutil.ReadFile(CPU_PATH + "/" + cpu_d.Name() + "/topology/thread_siblings_list")
			if err != nil {
				return err, nil
			}

			// Clean up the data
			dat := strings.TrimSpace(string(tsiblings))

			//log.Debug("CPU %s has thread(s) : %s", cpu_d.Name(), dat)
			log.WithFields(log.Fields{
				"cpu_id":  cpu_d.Name(),
				"threads": dat,
			}).Debug("CPU threads")

			id := strings.Split(cpu_d.Name(), "cpu")[1]
			id_i, err := strconv.Atoi(id)
			if err != nil {
				log.Error("Cannot convert CPU ID to int")
				return err, nil
			}

			cpus = append(cpus, CPU{Name: cpu_d.Name(), Id: id_i, ThreadsList: dat})
		}
	}

	log.WithFields(log.Fields{
		"topology": cpus,
	}).Debug("CPU Topology")
	return nil, cpus
}

// Activate the bit `bit` in the byte array `buf`. Bytes are stored in little-endian order.
func setBit(buf []byte, bit uint64) {
	idx := bit / 8 // 23 / 8 = 2
	pos := bit - idx*8
	log.WithFields(log.Fields{
		"bit":      bit,
		"byte":     idx,
		"position": pos,
	}).Debug("Setting bit in bitmap")
	buf[idx] = 1 << pos
}

// Pinning Guest (virDomain) Virtual CPU on Hypervisor CPU threads
func pinGuestToCPUThreads(d libvirt.VirDomain, countHostCpus uint16, countGuestCpus uint16, cpuTopo []CPU) {
	var vproc uint
	var cpumap []byte

	vproc = 0
	for i := countHostCpus - countGuestCpus; i < countHostCpus; i++ {
		threadlist := cpuTopo[i].ThreadsList

		log.WithFields(log.Fields{
			"vcpu":    vproc,
			"threads": threadlist,
		}).Info("Pinning VCPU on threads")

		cpumap = make([]byte, 6)
		for _, sbit := range strings.Split(threadlist, ",") {
			bit, _ := strconv.ParseUint(sbit, 10, 64)
			setBit(cpumap, bit)
		}
		d.PinVcpu(vproc, cpumap, 6)
		vproc++
	}
}

func main() {
	// Args
	var debug = flag.Bool("debug", false, "Debug mode")
	flag.Parse()

	// Variables
	var domains []libvirt.VirDomain

	// Setup logging
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	logLevel := log.InfoLevel
	if *debug {
		logLevel = log.DebugLevel
	}
	log.SetLevel(logLevel)

	// Get the CPU topology and sort it by CPU ID
	err, cpuTopology := get_cpu_topo()
	if err != nil {
		log.Fatal(err)
	}
	sort.Sort(ById(cpuTopology))

	// Connect to Qemu using Libvirt
	conn, err := libvirt.NewVirConnection("qemu:///system")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.CloseConnection()

	// Get Hypervisor CPUs info
	nodeInfo, _ := conn.GetNodeInfo()
	countCpus := nodeInfo.GetCPUs()

	// Get list of active domains
	flags := uint32(libvirt.VIR_CONNECT_LIST_DOMAINS_ACTIVE)
	domains, err = conn.ListAllDomains(flags)
	if err != nil {
		log.Fatal(err)
	}

	// Looping into active domains and pinning CPUs based on CPU Threads
	for _, domain := range domains {
		//id, _ := domain.GetID()
		name, _ := domain.GetName()
		log.WithFields(log.Fields{
			"name": name,
		}).Info("Working on domain")
		info, _ := domain.GetInfo()
		domainVcpus := info.GetNrVirtCpu()
		pinGuestToCPUThreads(domain, uint16(countCpus), uint16(domainVcpus), cpuTopology)
	}
}
