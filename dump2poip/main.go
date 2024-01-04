package main

import (
	"flag"
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcap"
	"github.com/gopacket/gopacket/pcapgo"
)

var (
	outFile   = flag.String("o", "file://output.pcap", "output pcap uri. can be file://, tcp://, tcplisten:// udp:// or unix://")
	device    = flag.String("i", "lo", "input device")
	bpf       = flag.String("b", "", "bpf filter")
	bigEndian = flag.Bool("be", false, "big endian")
)

var deduper = make(map[uint64]struct{})
var capacity = 10

// FNV1A is a very fast hashing function, mainly used for de-duplication
func FNV1A(input []byte) uint64 {
	var hash uint64 = 0xcbf29ce484222325
	var fnvPrime uint64 = 0x100000001b3
	for _, b := range input {
		hash ^= uint64(b)
		hash *= fnvPrime
	}
	return hash
}

func dedup(i uint64) bool {
	if _, ok := deduper[i]; ok {
		return true
	}
	if len(deduper) >= capacity {
		// remove oldest
		for k := range deduper {
			delete(deduper, k)
			break
		}
	}
	deduper[i] = struct{}{}
	return false
}

func main() {
	// open a new pcap file
	flag.Parse()

	outW, err := NewUriWriter().GetWriter(*outFile)
	if err != nil {
		log.Fatal(err)
	}

	var writer *pcapgo.Writer
	if *bigEndian {
		writer = pcapgo.NewWriterBigEndian(outW)
	} else {
		writer = pcapgo.NewWriter(outW)
	}

	handle, err := pcap.OpenLive(*device, 65535, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	if err := handle.SetBPFFilter(*bpf); err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	packetSrc := gopacket.NewPacketSource(handle, handle.LinkType()).Packets()
	writer.WriteFileHeader(65535, layers.LinkTypeEthernet)

	MAC, _ := net.ParseMAC("01:02:03:04:05:06")
	options := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	buffer := gopacket.NewSerializeBuffer()
	for {
		nextPacket := <-packetSrc
		data := nextPacket.Data()
		if dedup(FNV1A(data)) {
			continue
		}

		// Create the Ethernet layer
		var ethernetLayer gopacket.SerializableLayer
		myLayers := nextPacket.Layers()

		// if the first layer is not Ethernet, we need to prepend the Ethernet layer
		if nextPacket.Layers()[0].LayerType() == layers.LayerTypeIPv4 {
			ethernetLayer = &layers.Ethernet{
				SrcMAC:       MAC,
				DstMAC:       MAC,
				EthernetType: layers.EthernetTypeIPv4,
			}
		}
		if nextPacket.Layers()[0].LayerType() == layers.LayerTypeIPv6 {
			ethernetLayer = &layers.Ethernet{
				SrcMAC:       MAC,
				DstMAC:       MAC,
				EthernetType: layers.EthernetTypeIPv6,
			}
		}

		data = []byte{}
		for i := 0; i < len(myLayers); i++ {
			data = append(data, myLayers[i].LayerContents()...)
		}

		if ethernetLayer == nil {
			gopacket.SerializeLayers(buffer, options,
				gopacket.Payload(data),
			)
		} else {
			gopacket.SerializeLayers(buffer, options,
				ethernetLayer,
				gopacket.Payload(data),
			)
		}

		nextPacket = gopacket.NewPacket(buffer.Bytes(), layers.LinkTypeEthernet, gopacket.Default)

		packetHeader := gopacket.CaptureInfo{
			CaptureLength: len(nextPacket.Data()),
			Length:        len(nextPacket.Data()),
			Timestamp:     time.Now(),
		}
		if err := writer.WritePacket(packetHeader, nextPacket.Data()); err != nil {
			slog.Warn("error writing packet: ", "error", err)
		}
	}
}
