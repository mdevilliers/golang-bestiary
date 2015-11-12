package vagrant

import (
	"errors"
	"sync"
	"testing"

	"github.com/jonboulle/clockwork"
	"golang.org/x/net/context"
)

type MockJsonAccessor struct {
	eras []MockJsonEra
	era  int
}

type MockJsonEra struct {
	errorToReturn   error
	contentToReturn string
	url             string
}

func (m *MockJsonAccessor) MockJsonGetter(url string) ([]byte, error) {

	current := m.eras[m.era]
	m.roll()

	if current.errorToReturn != nil {
		return nil, current.errorToReturn
	}

	return []byte(current.contentToReturn), nil
}

func (m *MockJsonAccessor) roll() {

	m.era++
	if m.era == len(m.eras) {
		m.era = 0
	}
}

func MockifyWatcher(watcher *VagrantShareRemoteWatcher, mock *MockJsonAccessor, clock clockwork.Clock) {

	watcher.jsonGetter = mock.MockJsonGetter
	watcher.clock = clock
}

func TestBasicUsage(t *testing.T) {

	var wg sync.WaitGroup

	watcher := NewWatcher("test-url-does-not-exist")

	MockifyWatcher(watcher, &MockJsonAccessor{
		eras: []MockJsonEra{
			MockJsonEra{errorToReturn: errors.New("!")}},
	}, clockwork.NewRealClock())

	ctx, cancel := context.WithCancel(context.Background())

	go func() {

		select {
		case <-watcher.Updated:

			cancel()
			wg.Done()

		case <-watcher.OnError:

			cancel()
			wg.Done()
		}

	}()

	wg.Add(1)
	watcher.Watch(ctx)

	wg.Wait()

	if ctx.Err() != context.Canceled {
		t.Error("Should have been cancelled!")
	}
}

var fixture1 string = "{\"name\": \"test\",\"description\": \"description\",\"versions\": [{\"version\": \"201511.1111.1234\"} ]}"
var fixture2 string = "{\"name\": \"test\",\"description\": \"description\",\"versions\": [{\"version\": \"201511.2222.5678\"} ]}"

func TestWatcherEmitsOnChange(t *testing.T) {

	var wg sync.WaitGroup

	watcher := NewWatcher("test-url-does-not-exist")
	clock := clockwork.NewFakeClock()
	count := 0

	MockifyWatcher(watcher, &MockJsonAccessor{
		eras: []MockJsonEra{
			MockJsonEra{contentToReturn: fixture1},
			MockJsonEra{contentToReturn: fixture1},
			MockJsonEra{contentToReturn: fixture1},
			MockJsonEra{contentToReturn: fixture1},
			MockJsonEra{contentToReturn: fixture2}},
	}, clock)

	ctx, _ := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-watcher.Updated:
				count++
				wg.Done()

			case <-watcher.OnError:

				t.Error("No errors should have been returned.")
			}
		}
	}()

	wg.Add(1)

	watcher.Watch(ctx)

	wg.Wait()

	clock.Advance(watcher.periodInSeconds)
	wg.Add(1)

	wg.Wait()

	if count != 2 {
		t.Error("Should have received 2 events")
	}
}

func TestVersionsAreSortedCorrectly(t *testing.T) {

	descriptor := &VagrantBoxDescripter{
		Versions: Versions{
			Version{Version: "3"},
			Version{Version: "1"},
			Version{Version: "2"},
		},
	}

	descriptor.Versions.Sort()

	if descriptor.Versions[0].Version != "1" {
		t.Error("Versions not sorted correctly.")
	}
	if descriptor.Versions[1].Version != "2" {
		t.Error("Versions not sorted correctly.")
	}
	if descriptor.Versions[2].Version != "3" {
		t.Error("Versions not sorted correctly.")
	}
}
