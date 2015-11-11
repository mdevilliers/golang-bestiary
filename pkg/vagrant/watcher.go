package vagrant

import (
	"encoding/json"
	"golang.org/x/net/context"
	"hash/crc64"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"
)

type VagrantShareRemoteWatcher struct {
	url             string
	Updated         chan *VagrantShareDescripter
	periodInSeconds time.Duration
}

func NewWatcher(url string) *VagrantShareRemoteWatcher {
	return &VagrantShareRemoteWatcher{
		url:             url,
		Updated:         make(chan *VagrantShareDescripter),
		periodInSeconds: time.Second * 60,
	}
}

func (v *VagrantShareRemoteWatcher) Watch(ctx context.Context) {
	go v.loop(ctx)
}

func (v *VagrantShareRemoteWatcher) loop(ctx context.Context) {

	table := crc64.MakeTable(crc64.ECMA)
	var lastCheckSum uint64 = 0

	for {

		checkSum, descriptor, err := v.downloadJson(table)

		if err != nil {
			log.Fatalln(err.Error())
		} else {
			log.Print(checkSum)
			if lastCheckSum != checkSum {
				v.Updated <- descriptor
				lastCheckSum = checkSum
			}
		}

		select {

		case <-ctx.Done():
			return

		case <-time.After(v.periodInSeconds):

		}
	}
}

func (v *VagrantShareRemoteWatcher) downloadJson(table *crc64.Table) (uint64, *VagrantShareDescripter, error) {

	response, err := http.Get(v.url)

	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return 0, nil, err
	}
	checksum64 := crc64.Checksum(body, table)

	var descripter VagrantShareDescripter
	json.Unmarshal(body, &descripter)
	return checksum64, &descripter, nil
}

type VagrantShareDescripter struct {
	Name        string
	Description string
	Versions    Versions
}

type Versions []Version

func (a Versions) Len() int           { return len(a) }
func (a Versions) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Versions) Less(i, j int) bool { return a[i].Version < a[j].Version }

func (v Versions) Sort() {
	sort.Sort(Versions(v))
}

type Version struct {
	Version   string
	Providers []Provider
}

type Provider struct {
	Name         string
	Url          string
	CheckSumType string `json:"checksum_type"`
	CheckSum     string
}
