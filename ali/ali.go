package ali

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

func GetRecordsId(rr, domainName, id, key string) (string, error) {
	client, err := alidns.NewClientWithAccessKey("cn-qingdao", id, key)
	if err != nil {
		return "", err
	}
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = domainName
	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	for _, v := range response.DomainRecords.Record {
		if v.RR == rr {
			return v.RecordId, nil
		}
	}
	return "", nil
}

func AddRecord(rr, domain, value, id, key string) (string, error) {
	if rr == "" {
		rr = "@"
	}
	client, err := alidns.NewClientWithAccessKey("cn-qingdao", id, key)
	if err != nil {
		return "", err
	}
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"

	request.Value = value
	request.Type = "AAAA"
	request.RR = rr
	request.DomainName = domain

	response, err := client.AddDomainRecord(request)
	return response.RecordId, err
}

func UpdateRecord(rr, recordId, value, id, key string) error {
	client, err := alidns.NewClientWithAccessKey("cn-qingdao", id, key)
	if err != nil {
		return err
	}
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.RecordId = recordId
	request.RR = rr
	request.Type = "AAAA"
	request.Value = value

	response, err := client.UpdateDomainRecord(request)
	if err != nil || response.RecordId != recordId {
		return err
	}
	return nil
}
