# go_pprof_test
* simple go pprof sample

# build
* go build .

# execute
* ./go_pprof_test
  * output "cpu_<now_unixtime>.pprof"

# profiling
* go tool pprof -http=:8080 path/to/cpu.prof

