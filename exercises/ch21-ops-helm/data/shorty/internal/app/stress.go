package app

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// The /stress routes exist only in the chapter 21 copy of shorty. They make
// resource limits visible on purpose: memory that is never freed (to reach an
// OOM kill) and CPU work per request (to trigger throttling and the HPA).
// They are registered only when SHORTY_STRESS=on, never in a real service.

var (
	hogMu sync.Mutex
	hog   [][]byte // memory held by POST /stress/memory, never released
)

func registerStress(mux *http.ServeMux) {
	mux.HandleFunc("POST /stress/memory", stressMemory)
	mux.HandleFunc("GET /stress/cpu", stressCPU)
}

// queryInt reads ?name=N and checks 1 <= N <= upper.
func queryInt(r *http.Request, name string, upper int) (int, error) {
	n, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil || n < 1 || n > upper {
		return 0, fmt.Errorf("%s must be an integer between 1 and %d", name, upper)
	}
	return n, nil
}

// stressMemory is POST /stress/memory?mb=N: allocate N MiB, write to every
// page so the kernel really backs it, and keep it until the process dies.
func stressMemory(w http.ResponseWriter, r *http.Request) {
	mb, err := queryInt(r, "mb", 1024)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	buf := make([]byte, mb<<20)
	for i := 0; i < len(buf); i += 4096 {
		buf[i] = 1
	}
	hogMu.Lock()
	hog = append(hog, buf)
	total := 0
	for _, b := range hog {
		total += len(b)
	}
	hogMu.Unlock()
	fmt.Fprintf(w, "holding %d MiB\n", total>>20)
}

// stressCPU is GET /stress/cpu?work=N: a fixed amount of arithmetic, N
// million steps (a few milliseconds of one core each), and the wall-clock time
// it took. Under a CPU limit the work stays the same and the time grows: that
// difference is throttling.
func stressCPU(w http.ResponseWriter, r *http.Request) {
	work, err := queryInt(r, "work", 1000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	start := time.Now()
	x := uint64(1)
	for range work * 1_000_000 {
		x = x*6364136223846793005 + 1442695040888963407
	}
	fmt.Fprintf(w, "work %d, wall %d ms, checksum %d\n", work, time.Since(start).Milliseconds(), x&0xff)
}
