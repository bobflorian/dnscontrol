package ns1

import (
	"encoding/json"
	"fmt"

	"net/http"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/miekg/dns/dnsutil"
	api "gopkg.in/ns1/ns1-go.v2/rest"
	apiModels "gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

type ns1Provider struct {
	client *api.Client
}

func init() {
	providers.RegisterDomainServiceProviderType("NS1", newClient)
}

func newClient(creds map[string]string, meta json.RawMessage) (providers.DNSServiceProvider, error) {
	key := creds["apikey"]
	if key == "" {
		return nil, fmt.Errorf("NS1 apikey must be provided")
	}
	client := api.NewClient(http.DefaultClient, api.SetAPIKey(key))
	return &ns1Provider{client: client}, nil
}

func (n *ns1Provider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, _, err := n.client.Zones.Get(domain)
	if err != nil {
		return nil, err
	}
	return models.StringsToNameservers(zone.DNSServers), nil
}
func (n *ns1Provider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	zone, _, err := n.client.Zones.Get(dc.Name)
	if err != nil {
		return nil, err
	}
	found := []*models.RecordConfig{}
	for _, rec := range zone.Records {
		found = append(found, ns1ToRecords(rec, dc.Name)...)
	}
	for _, rec := range dc.Records {
		rec.InlineMXPriority()
	}
	differ := diff.New(dc)
	_, create, del, modify := differ.IncrementalDiff(found)

	for _, c := range create {
		fmt.Println(c)
	}
	for _, d := range del {
		fmt.Println(d)
	}
	for _, m := range modify {
		fmt.Println(m)
	}
	return nil, nil
}

func ns1ToRecords(zr *apiModels.ZoneRecord, origin string) []*models.RecordConfig {
	recs := []*models.RecordConfig{}
	for _, ans := range zr.ShortAns {
		rec := &models.RecordConfig{
			Name:     dnsutil.TrimDomainName(zr.Domain, origin),
			NameFQDN: zr.Domain,
			Type:     zr.Type,
			TTL:      uint32(zr.TTL),
			Target:   ans,
		}
		recs = append(recs, rec)
	}
	return recs
}
