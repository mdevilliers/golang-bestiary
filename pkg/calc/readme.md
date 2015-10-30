

```
go build -buildmode=c-shared -o calc-shared.a calc.go
gcc -o calc_golang_to_c calc_c_to_go.c ~/go/src/github.com/mdevilliers/golang-bestiary/pkg/calc/calc-shared.a
```
