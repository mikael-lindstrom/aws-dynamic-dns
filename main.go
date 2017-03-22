package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

const ipProvider = "http://checkip.amazonaws.com"
const hostedZoneIdEnvKey = "HOSTED_ZONE_ID"
const domainEnvKey = "DNS_NAME"

type Route53 struct {
	svc            *route53.Route53
	domain, zoneId string
}

func New() (r53 *Route53) {
	zoneId, ok := os.LookupEnv(hostedZoneIdEnvKey)
	if !ok {
		exitOnError(fmt.Errorf("Environment variable '%s' not set\n", hostedZoneIdEnvKey))
	}

	domain, ok := os.LookupEnv(domainEnvKey)
	if !ok {
		exitOnError(fmt.Errorf("Environment variable '%s' not set\n", domainEnvKey))
	}

	session := session.Must(session.NewSession())
	svc := route53.New(session)
	return &Route53{
		svc:    svc,
		zoneId: zoneId,
		domain: domain,
	}
}

func exitOnError(err error) {
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func getPublicIp() (publicIp string) {
	response, err := http.Get(ipProvider)
	exitOnError(err)

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	exitOnError(err)

	return strings.TrimSpace(string(body))
}

func (r53 *Route53) getRoute53RecordIp() (route53Ip string) {
	params := &route53.TestDNSAnswerInput{
		HostedZoneId: aws.String(r53.zoneId),
		RecordName:   aws.String(r53.domain),
		RecordType:   aws.String("A"),
	}
	resp, err := r53.svc.TestDNSAnswer(params)
	exitOnError(err)

	if resp.RecordData == nil {
		// No record currently exists
		return ""
	}
	return aws.StringValue(resp.RecordData[0])
}

func (r53 *Route53) updateRoute53RecordIp(publicIp string) {
	recordSet := &route53.ResourceRecordSet{
		Name: aws.String(r53.domain),
		Type: aws.String("A"),
		TTL:  aws.Int64(300),
		ResourceRecords: []*route53.ResourceRecord{
			{
				Value: aws.String(publicIp),
			},
		},
	}

	changeBatch := &route53.ChangeBatch{
		Comment: aws.String("aws-dynamic-dns"),
		Changes: []*route53.Change{
			{
				Action:            aws.String("UPSERT"),
				ResourceRecordSet: recordSet,
			},
		},
	}

	params := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(r53.zoneId),
		ChangeBatch:  changeBatch,
	}

	fmt.Println(params)

	_, err := r53.svc.ChangeResourceRecordSets(params)
	exitOnError(err)
}

func main() {
	r53 := New()

	fmt.Printf("Using HostedZoneId '%s' and domain '%s'\n", r53.zoneId, r53.domain)

	route53Ip := r53.getRoute53RecordIp()
	publicIp := getPublicIp()

	if route53Ip != publicIp {
		fmt.Printf("Route53 record IP '%s' does not match public IP '%s', updating record\n", route53Ip, publicIp)
		r53.updateRoute53RecordIp(publicIp)
		fmt.Printf("Route53 record '%s' updated to IP '%s' (not waiting for record to be in sync)\n", r53.domain, publicIp)
	} else {
		fmt.Printf("Route53 record '%s' already matches IP '%s'\n", r53.domain, publicIp)
	}
}
