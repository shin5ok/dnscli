package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shin5ok/dnscli/internal/clouddns"
)

var defaultConfig = map[string]string{
	"domain":  os.Getenv("DNS_DOMAIN"),
	"zone":    os.Getenv("DNS_ZONE"),
	"project": os.Getenv("GOOGLE_CLOUD_PROJECT"),
}

type config struct {
	Project string `json:"project"`
	Domain  string `json:"domain"`
	Zone    string `json:"zone"`
}

func envinfo(c config) {
	data, err := json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}

type DNSMain struct {
	Client clouddns.Recorder
}

func usageShow() {
	usage := `
View configuration:
  $ clouddns -env
Usage:
  $ clouddns -zone ZONE_NAME -domain DOMAIN_NAME -key RECORD -data DATA
Example:
  $ clouddns -zone w-uname-link -domain w.uname.link. -key api.w.uname.link. -data 10.10.10.10
`
	fmt.Println(usage)
}

func main() {
	data := flag.String("data", "", "")
	key := flag.String("key", "", "")
	domain := flag.String("domain", defaultConfig["domain"], `ig: example.com`)
	zone := flag.String("zone", defaultConfig["zone"], `ig: exapmple-com`)
	rtype := flag.String("type", "A", "")
	projectId := flag.String("project", defaultConfig["project"], "")
	ttl := flag.Int64("ttl", 60, "")
	env := flag.Bool("env", false, "")
	help := flag.Bool("help", false, "")
	flag.Parse()

	if *env {
		c := config{
			Domain:  *domain,
			Project: *projectId,
			Zone:    *zone,
		}
		envinfo(c)
		return
	}

	if *zone == "" || *help {
		usageShow()
		return
	}

	rr := clouddns.Record{
		RType: *rtype,
		RData: []string{*data},
		RKey:  *key,
		TTL:   int(*ttl),
	}

	zoneInfo := clouddns.ZoneInfo{
		Domain:      *domain,
		ProjectId:   *projectId,
		ManagedZone: *zone,
	}

	newRr, err := run(&zoneInfo, rr, *key)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonResult, err := json.Marshal(&newRr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Updated: %v\n", string(jsonResult))

}

func run(zoneInfo clouddns.Recorder, rr clouddns.Record, key string) (clouddns.Record, error) {

	v := DNSMain{
		Client: zoneInfo,
	}

	_, err := v.Client.Get(key, "A")
	if errors.Is(err, clouddns.ErrNotFound) {
		err := v.Client.Create(&rr)
		if err != nil {
			return clouddns.Record{}, err
		}
	} else {
		err := v.Client.Set(&rr)
		if err != nil {
			return clouddns.Record{}, err
		}
	}

	return rr, nil
}
