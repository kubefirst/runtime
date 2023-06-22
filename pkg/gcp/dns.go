/*
Copyright (C) 2021-2023, Kubefirst

This program is licensed under MIT.
See the LICENSE file for more details.
*/
package gcp

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/kubefirst/runtime/pkg/dns"
	"github.com/rs/zerolog/log"
	googleDNS "google.golang.org/api/dns/v1"
)

// TestHostedZoneLiveness checks DNS for the liveness test record
func (conf *GCPConfiguration) TestHostedZoneLiveness(hostedZoneName string) bool {
	recordName := fmt.Sprintf("kubefirst-liveness.%s.", hostedZoneName)
	recordValue := "domain record propagated"

	dnsService, err := googleDNS.NewService(conf.Context)
	if err != nil {
		log.Error().Msgf("error creating google dns client: %s", err)
		return false
	}

	zones, err := dnsService.ManagedZones.List(conf.Project).Do()
	if err != nil {
		log.Error().Msgf("error listing google dns zones: %s", err)
		return false
	}

	var zoneName string

	for _, zone := range zones.ManagedZones {
		if strings.Contains(zone.DnsName, hostedZoneName) {
			zoneName = zone.Name
		}
	}
	if zoneName == "" {
		log.Error().Msgf("could not find zone %s", hostedZoneName)
		return false
	}

	zone, err := dnsService.ManagedZones.Get(conf.Project, zoneName).Do()
	if err != nil {
		log.Error().Msgf("error getting google dns zone %s: %s", hostedZoneName, err)
		return false
	}

	log.Info().Msgf("checking to see if record %s exists", recordName)
	log.Info().Msgf("recordName %s", recordName)

	// check for existing record
	records, err := dnsService.ResourceRecordSets.List(conf.Project, zone.Name).Do()
	if err != nil {
		log.Error().Msgf("error listing google dns zone records: %s", err)
		return false
	}

	for _, r := range records.Rrsets {
		if r.Name == recordName {
			log.Info().Msg("domain record found")
			return true
		}
	}

	// create record if it does not exist
	stat, err := dnsService.ResourceRecordSets.Create(conf.Project, zone.Name, &googleDNS.ResourceRecordSet{
		Name:    recordName,
		Rrdatas: []string{recordValue},
		Ttl:     10,
		Type:    "TXT",
	}).Do()
	if err != nil {
		log.Error().Msgf("error creating google dns zone record: %s", err)
		return false
	}
	log.Info().Msgf("record creation status is %v", stat.HTTPStatusCode)

	count := 0
	// todo need to exit after n number of minutes and tell them to check ns records
	// todo this logic sucks
	for count <= 100 {
		count++

		log.Info().Msgf("%s", recordName)
		ips, err := net.LookupTXT(recordName)
		if err != nil {
			ips, err = dns.BackupResolver.LookupTXT(context.Background(), recordName)
		}

		log.Info().Msgf("%s", ips)

		if err != nil {
			log.Warn().Msgf("could not get record name %s - waiting 10 seconds and trying again: \nerror: %s", recordName, err)
			time.Sleep(10 * time.Second)
		} else {
			for _, ip := range ips {
				// todo check ip against route53RecordValue in some capacity so we can pivot the value for testing
				log.Info().Msgf("%s. in TXT record value: %s\n", recordName, ip)
				count = 101
			}
		}
		if count == 100 {
			log.Error().Msg("unable to resolve hosted zone dns record. please check your domain registrar")
			return false
		}
	}
	return true
}
