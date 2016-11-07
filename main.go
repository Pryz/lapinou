package main

import (
	"flag"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rgbkrk/libvirt-go"
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
func pinGuestToCPUThreads(d libvirt.VirDomain, countHostCpus uint32, countGuestCpus uint32, cpuTopo []CPU, pCount uint32) {
	var vproc uint
	var cpuMap []uint32

	idx := countHostCpus - pCount

	vproc = 0
	for i := idx - countGuestCpus; i < idx; i++ {
		threadList := cpuTopo[i].ThreadsList

		log.WithFields(log.Fields{
			"vcpu":    vproc,
			"threads": threadList,
			"cpu":     cpuTopo[i].Id,
		}).Info("Pinning VCPU on threads")

		cpuMap = make([]uint32, 2)
		for i, sbit := range strings.Split(threadList, ",") {
			bit, _ := strconv.ParseUint(sbit, 10, 32)
			cpuMap[i] = uint32(bit)
		}
		d.PinVcpu(vproc, cpuMap, countHostCpus)
		vproc++
	}
}

// Apply the pinning strategy against all domains passed in parameter
func doPinning(ds []libvirt.VirDomain, hostCpus uint32, cpuTopology []CPU) bool {
	var totalVcpus uint16
	var pinnedCount uint32

	// Count number of provisioned VCPU(s)
	totalVcpus = 0
	for _, domain := range ds {
		info, _ := domain.GetInfo()
		domainVcpus := info.GetNrVirtCpu()
		totalVcpus += domainVcpus
	}
	if uint32(totalVcpus) > hostCpus {
		log.Info("Not enough CPU(s) to apply pinning on all provisioned Domain(s). Skipping")
		return false
	}
	log.Info("Enough CPUs to apply pinning on provisioned VCPUs")

	// Looping into active domains and pinning CPUs based on CPU Threads
	pinnedCount = 0
	for _, domain := range ds {
		name, _ := domain.GetName()
		log.WithFields(log.Fields{
			"name": name,
		}).Info("Working on domain")
		info, _ := domain.GetInfo()
		domainVcpus := info.GetNrVirtCpu()
		// TODO: We should be able to define and use different strategy here
		// TODO: Should we use a go routine here ?
		pinGuestToCPUThreads(domain, hostCpus, uint32(domainVcpus), cpuTopology, pinnedCount)
		pinnedCount += uint32(domainVcpus)
	}
	return true
}

func main() {
	// Args
	var debug = flag.Bool("debug", false, "Debug mode")
	var cli = flag.Bool("cli", false, "CLI Mode. Otherwise Lapinou will be 'daemonized'")
	var jsonlog = flag.Bool("jsonlog", false, "Output logs in JSON instead of text")
	flag.Parse()

	// Variables
	var domains []libvirt.VirDomain

	// Setup logging
	if *jsonlog {
		log.SetFormatter(&log.JSONFormatter{})
	}
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
	// TODO: The connection URI should be configurable
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
	log.WithFields(log.Fields{
		"cpus": countCpus,
	}).Debug("Node Info")
	log.WithFields(log.Fields{
		"count": len(domains),
	}).Info("Domains founds")

	if *cli {
		// Run the pinning stategy once and exit. Easy to use within a crontab
		doPinning(domains, countCpus, cpuTopology)
	} else {
		// "Daemonized" execution
		for {
			if doPinning(domains, countCpus, cpuTopology) {
				log.Info("CPU Pinning successful. Will check again in 5min.")
			} else {
				log.Info("CPU Pinning failed. Will retry again in 5min.")
			}
			time.Sleep(5 * time.Minute)
		}
	}
}
