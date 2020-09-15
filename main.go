package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
)

const defaultSource = "https://github.com/v2fly/geoip/raw/release/geoip.dat"

var (
	source  = flag.String("source", defaultSource, "IP data file.")
	country = flag.String("country", "CN", "Country.")
	ipv4Out = flag.String("ipv4_out", "", "IPv4 address output file.")
	ipv6Out = flag.String("ipv6_out", "", "IPv6 address output file.")
)

func loadIP(filename, country string) ([]*CIDR, error) {
	var err error
	var geoipBytes []byte
	if _, err = url.ParseRequestURI(filename); err == nil {
		resp, err := http.Get(filename)
		if err != nil {
			return nil, err
		}
		geoipBytes, err = ioutil.ReadAll(resp.Body)
	} else {
		geoipBytes, err = ioutil.ReadFile(filename)
	}
	if err != nil {
		return nil, errors.New("failed to open file: " + filename)
	}
	var geoipList GeoIPList
	if err := proto.Unmarshal(geoipBytes, &geoipList); err != nil {
		return nil, err
	}

	for _, geoip := range geoipList.Entry {
		if geoip.CountryCode == country {
			return geoip.Cidr, nil
		}
	}

	return nil, errors.New("country not found: " + country)
}

func loadSite(filename, country string) ([]*Domain, error) {
	geositeBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New("failed to open file: " + filename)
	}
	var geositeList GeoSiteList
	if err := proto.Unmarshal(geositeBytes, &geositeList); err != nil {
		return nil, err
	}

	for _, site := range geositeList.Entry {
		if site.CountryCode == country {
			return site.Domain, nil
		}
	}

	return nil, errors.New("country not found: " + country)
}

func main() {
	flag.Parse()
	if *ipv4Out == "" && *ipv6Out == "" {
		flag.PrintDefaults()
		return
	}
	ips, err := loadIP(*source, *country)
	if err != nil {
		log.Fatal(err)
	}
	var ipv4File, ipv6File *os.File
	if *ipv4Out != "" {
		ipv4File, err = os.Create(*ipv4Out)
		if err != nil {
			log.Fatal(err)
		}
		defer ipv4File.Close()
	}
	if *ipv6Out != "" {
		ipv6File, err = os.Create(*ipv6Out)
		if err != nil {
			log.Fatal(err)
		}
		defer ipv6File.Close()
	}

	for _, ip := range ips {
		addr := net.IP(ip.Ip)
		addrString := fmt.Sprintf("%s/%d\n", addr.String(), ip.Prefix)
		if len(ip.Ip) == 4 && ipv4File != nil {
			ipv4File.WriteString(addrString)
		} else if len(ip.Ip) == 16 && ipv6File != nil {
			ipv6File.WriteString(addrString)
		}
	}
}
