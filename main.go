package main

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
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
	if len(os.Args) < 4 {
		println(os.Args[0] + " source country output")
		os.Exit(1)
		return
	}
	ips, err := loadIP(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Create(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	for _, ip := range ips {
		file.WriteString(fmt.Sprintf("%d.%d.%d.%d/%d\n", ip.Ip[0], ip.Ip[1], ip.Ip[2], ip.Ip[3], ip.Prefix))
	}
}
