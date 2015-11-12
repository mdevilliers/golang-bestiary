package vagrant

import (
	"encoding/json"
	"github.com/jonboulle/clockwork"
	"golang.org/x/net/context"
	"hash/crc64"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
)

type JsonDownloader func(url string) ([]byte, error)

type VagrantShareRemoteWatcher struct {
	url             string
	Updated         chan *VagrantBoxDescripter
	OnError         chan error
	periodInSeconds time.Duration
	jsonGetter      JsonDownloader
	clock           clockwork.Clock
}

func NewWatcher(url string) *VagrantShareRemoteWatcher {
	return &VagrantShareRemoteWatcher{
		url:             url,
		Updated:         make(chan *VagrantBoxDescripter),
		OnError:         make(chan error),
		periodInSeconds: time.Second * 60,
		jsonGetter:      jsonGetter,
		clock:           clockwork.NewRealClock(),
	}
}

func (v *VagrantShareRemoteWatcher) Watch(ctx context.Context) {
	go v.loop(ctx)
}

func (v *VagrantShareRemoteWatcher) loop(ctx context.Context) {

	table := crc64.MakeTable(crc64.ECMA)
	var lastCheckSum uint64 = 0

	for {

		go func() {
			checkSum, descriptor, err := v.downloadJson(table)

			if err != nil {
				v.OnError <- err
			} else {

				if lastCheckSum != checkSum {

					v.Updated <- descriptor
					lastCheckSum = checkSum
				}
			}
		}()

		select {

		case <-ctx.Done():
			return

		case <-v.clock.After(v.periodInSeconds):
			// continue
		}
	}
}

func (v *VagrantShareRemoteWatcher) downloadJson(table *crc64.Table) (uint64, *VagrantBoxDescripter, error) {

	body, err := v.jsonGetter(v.url)

	if err != nil {
		return 0, nil, err
	}
	checksum64 := crc64.Checksum(body, table)

	var descripter VagrantBoxDescripter

	json.Unmarshal(body, &descripter)

	return checksum64, &descripter, nil
}

func jsonGetter(url string) ([]byte, error) {

	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

type VagrantBoxDescripter struct {
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
