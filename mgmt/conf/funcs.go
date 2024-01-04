package conf

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

var NFT_CHAIN_NAME = "aio-gw-chain"

func doesChainExist(chainName string) (bool, error) {
	nftConn := nftables.Conn{}
	a, err := nftConn.ListChains()
	if err != nil {
		log.Errorf("Error listing iptables chains: %s", err)
		return false, err
	}
	for _, chain := range a {
		log.Warnf("Checking chain %s", chain.Name)
		if chain.Name == chainName {
			return true, nil
		}
	}
	return false, nil
}

func createMasqChain(chainName string, family int) error {
	//nft 'add chain nat postrouting { type nat hook postrouting priority 100 ; }'
	// nft add rule nat postrouting masquerade
	tableFamily := nftables.TableFamily(nftables.TableFamilyIPv4)
	if family == 6 {
		tableFamily = nftables.TableFamily(nftables.TableFamilyIPv6)
	}
	nftConn := nftables.Conn{}
	myTable := nftables.Table{
		Name:   "nat",
		Family: tableFamily,
	}
	myPolicy := nftables.ChainPolicyAccept
	myChain := nftables.Chain{
		Name:     chainName,
		Table:    &myTable,
		Type:     nftables.ChainTypeNAT,
		Priority: nftables.ChainPriorityNATSource,
		Hooknum:  nftables.ChainHookPostrouting,
		Policy:   &myPolicy,
	}
	nftConn.AddTable(&myTable)
	nftConn.AddChain(&myChain)
	if err := nftConn.Flush(); err != nil {
		log.Errorf("failed to program with error: %+v\n", err)
		return err
	}
	nftConn.AddRule(&nftables.Rule{
		Table: &myTable,
		Chain: &myChain,
		Exprs: []expr.Any{

			&expr.Masq{},
		},
	})
	return nil
}

func createNatChain(chainName string, family int, ports []int, targetIP string, targetPort int) error {
	tableFamily := nftables.TableFamily(nftables.TableFamilyIPv4)
	if family == 6 {
		tableFamily = nftables.TableFamily(nftables.TableFamilyIPv6)
	}
	nftConn := nftables.Conn{}
	myTable := nftables.Table{
		Name:   "nat",
		Family: tableFamily,
	}
	myPolicy := nftables.ChainPolicyAccept
	myChain := nftables.Chain{
		Name:     chainName,
		Table:    &myTable,
		Type:     nftables.ChainTypeNAT,
		Priority: nftables.ChainPriorityNATDest,
		Hooknum:  nftables.ChainHookPrerouting,
		Policy:   &myPolicy,
	}

	nftConn.AddTable(&myTable)
	nftConn.AddChain(&myChain)
	if err := nftConn.Flush(); err != nil {
		log.Errorf("failed to program with error: %+v\n", err)
		return err
	}
	for _, port := range ports {
		log.Infof("Adding rule for port %d", port)

		myExpr := []expr.Any{
			// [ meta load iifname => reg 1 ]
			&expr.Meta{Key: expr.MetaKeyIIFNAME, Register: 1},
			// &expr.Cmp{
			// 	Op:       expr.CmpOpEq,
			// 	Register: 1,
			// 	Data:     []byte("eth0\x00"),
			// },
			// [ meta load l4proto => reg 1 ]
			&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
			// [ cmp eq reg 1 0x00000006 ]
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.IPPROTO_TCP},
			},
			// [ payload load 2b @ transport header + 2 => reg 1 ]
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2,
				Len:          2,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     binaryutil.BigEndian.PutUint16(uint16(port)),
			},
			&expr.Immediate{
				Register: 1,
				Data:     net.ParseIP(targetIP).To4(),
			},
			&expr.Immediate{
				Register: 2,
				Data:     binaryutil.BigEndian.PutUint16(uint16(targetPort)),
			},
			// [ nat dnat ip addr_min reg 1 addr_max reg 0 proto_min reg 2 proto_max reg 0 ]
			// &expr.NAT{
			// 	Type:        expr.NATTypeDestNAT,
			// 	Family:      unix.NFPROTO_IPV4,
			// 	RegAddrMin:  1,
			// 	RegProtoMin: 2,
			// },
			&expr.Redir{
				RegisterProtoMin: 2,
			},
		}

		nftConn.AddRule(
			&nftables.Rule{
				Table: &myTable,
				Chain: &myChain,
				Exprs: myExpr,
			},
		)
	}
	if err := nftConn.Flush(); err != nil {
		log.Errorf("failed to program with error: %+v\n", err)
		return err
	}
	return nil
}

func GetTagValue(myStruct interface{}, myField string, myTag string) string {
	t := reflect.TypeOf(myStruct)
	for i := 0; i < t.NumField(); i++ {
		// Get the field, returns https://golang.org/pkg/reflect/#StructField
		field := t.Field(i)
		if field.Name == myField {
			// Get the field tag value
			tag := field.Tag.Get(myTag)
			return tag
		}
	}

	return ""
}

func ApplyUpstreamSettings(upstreamSettings UpstreamSettings) error {
	log.Infof("Applying upstream settings %#v", upstreamSettings)
	if !upstreamSettings.Enabled {
		return nil
	}
	switch method := upstreamSettings.Method; method {
	case "dummy":
		return nil //todo
	case "sinkhole":
		return nil //todo
	case "tls_decryption":
		log.Infof("upstream is tls_decryption.. applying settings")
		// check for nft availability
		a, err := doesChainExist(NFT_CHAIN_NAME)
		if err != nil {
			log.Errorf("Error checking nft chain: %s", err)
			return err
		} else if a {
			log.Infof("nft chain %s already exists", NFT_CHAIN_NAME)
		} else {
			log.Infof("Creating nft chain %s", NFT_CHAIN_NAME)
			createNatChain(NFT_CHAIN_NAME, 4, upstreamSettings.TlsDecrpytionOptions.Ports, upstreamSettings.TlsDecrpytionOptions.Target.Ip, upstreamSettings.TlsDecrpytionOptions.Target.Port)
			createNatChain(NFT_CHAIN_NAME, 6, upstreamSettings.TlsDecrpytionOptions.Ports, upstreamSettings.TlsDecrpytionOptions.Target.Ip, upstreamSettings.TlsDecrpytionOptions.Target.Port)
			createMasqChain(NFT_CHAIN_NAME, 4)
			createMasqChain(NFT_CHAIN_NAME, 6)
		}

	default:
		return nil
	}
	return nil

}

// check current service settings, scrape them and apply the new config instead
func ApplyProxySettings(proxySettings ProxySettings) error {
	log.Warnf("Proxy function hasn't been implemented yet\n")
	return nil
}

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// make sure podman is up and running before calling this function. I'm not checking for it here
func getContainerNames(podmanPath string) []string {
	//podman ps --format {{.Names}}
	cmd := exec.Command(podmanPath, "ps", "-a", "--format", "{{.Names}}")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Errorf("Error getting container names: %s", err)
		return nil
	}
	containerNames := strings.Split(string(out), "\n")
	return containerNames
}

// check current service settings, scrape them and apply the new config instead
func ApplyServiceSettings(serviceSettings ServiceSettings) error {
	// check availablity of podman
	path, err := exec.LookPath("podman")
	if err != nil {
		log.Error("Podman binary not found")
		return err
	}
	log.Infof("podman path: %s", path)
	// start podman
	cmd := exec.Command(path, "run", "--rm", "--name", "my-container", "--net=host", "docker.io/hello-world")
	log.Infof("%v", cmd)
	err = cmd.Run()
	if err != nil {
		log.Errorf("Error running podman: %s", err)
		return err
	}
	cmd.Wait()
	log.Infof("podman ran hello world successfully")

	// get a list of all containers (running, stopped etc)
	containerNameList := getContainerNames(path)

	// (re)running containers based on their names
	for _, container := range serviceSettings.Containers {
		// don't do anything if the service is not enabled
		if !container.Enabled {
			continue
		}
		// checking if the container with the same name exist
		_, isFound := Find(containerNameList, container.Name)
		if isFound {
			// removing existing container and re-creating
			log.Infof("Removing container %s", container.Name)
			cmd = exec.Command(path, "rm", "-f", container.Name)
			err = cmd.Run()
			if err != nil {
				log.Errorf("Error removing container %s: %s", container.Name, err)
				return err
			}
			cmd.Wait()
			log.Infof("Removed container %s", container.Name)
		}
		log.Infof("Creating container %s", container.Name)
		// create container
		// podman run -d --name polar --net host mosajjal/polarproxy:latest -v -p 10443,80,443 --certhttp 1081 --pcapoveripconnect 127.0.0.1:57012 --cn Fortinet_CA_SSL
		runOptions := fmt.Sprintf("run -d --net host --name %s", container.Name)
		optionList := strings.Split(runOptions, " ")
		if len(container.PodmanOptions) > 0 {
			optionList = append(optionList, container.PodmanOptions...)
		}
		optionList = append(optionList, container.Image)
		if len(container.EntryOptions) > 0 {
			optionList = append(optionList, container.EntryOptions...)
		}
		cmd = exec.Command(path, optionList...)
		log.Infof("%v", optionList)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			log.Errorf("Error running container %s: %s", container.Name, err)
			return err
		}
		cmd.Wait()
		log.Infof("Created container %s", container.Name)
		// todo: run on boot
	}
	return nil
}
