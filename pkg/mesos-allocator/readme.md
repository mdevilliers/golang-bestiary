
build

```
g++ -c -lmesos -fpic -std=c++11 -o golang_allocator_module.o golang_allocator_module.cpp
gcc -shared -o libgolang_allocator_module.so golang_allocator_module.o

```

inspect

```
nm -gC ./libgolang_allocator_module.so 

```

dependancies

mesos
protobuf
boost
glog
curl-dev