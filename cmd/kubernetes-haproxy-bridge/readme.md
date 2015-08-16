

http://serverfault.com/questions/127491/haproxy-forward-to-a-different-web-server-based-on-uri

```
frontend http-in
    bind 10.254.23.225:80
    acl has_special_uri path_beg /special
    use_backend special_server if has_special_uri
    default_backend webfarm

backend webfarm
    balance roundrobin
    cookie SERVERID insert
    option httpchk HEAD /check.txt HTTP/1.0
    option httpclose
    option forwardfor
    server webA 10.254.23.4:80 cookie webA check
    server webB 10.248.23.128:80 cookie webB check

backend special_server
    balance roundrobin
    cookie SERVERID insert
    option httpchk HEAD /check.txt HTTP/1.0
    option httpclose
    option forwardfor
    server webC 10.0.0.1:80 cookie webC check
```

Kubernentes :
```
	(api.Service) {
 TypeMeta: (api.TypeMeta) {
  Kind: (string) "",
  APIVersion: (string) ""
 },
 ObjectMeta: (api.ObjectMeta) {
  Name: (string) (len=11) "service-one",
  GenerateName: (string) "",
  Namespace: (string) (len=7) "default",
  SelfLink: (string) (len=47) "/api/v1/namespaces/default/services/service-one",
  UID: (types.UID) (len=36) "a872e5b1-4451-11e5-ac27-080027aecac8",
  ResourceVersion: (string) (len=4) "2669",
  Generation: (int64) 0,
  CreationTimestamp: (util.Time) 2015-08-16 21:02:01 +0100 BST,
  DeletionTimestamp: (*util.Time)(<nil>),
  Labels: (map[string]string) (len=2) {
   (string) (len=15) "external/public": (string) (len=4) "true",
   (string) (len=4) "name": (string) (len=16) "service-one-node"
  },
  Annotations: (map[string]string) (len=4) {
   (string) (len=13) "external/name": (string) (len=11) "service-one",
   (string) (len=13) "external/port": (string) (len=4) "8080",
   (string) (len=12) "external/uri": (string) (len=13) "/path1/child/",
   (string) (len=13) "external/ytpe": (string) (len=4) "http"
  }
 },
 Spec: (api.ServiceSpec) {
  Ports: ([]api.ServicePort) (len=1 cap=1) {
   (api.ServicePort) {
    Name: (string) "",
    Protocol: (api.Protocol) (len=3) "TCP",
    Port: (int) 80,
    TargetPort: (util.IntOrString) 80,
    NodePort: (int) 30422
   }
  },
  Selector: (map[string]string) (len=1) {
   (string) (len=4) "name": (string) (len=16) "service-one-node"
  },
  ClusterIP: (string) (len=13) "10.100.20.151",
  Type: (api.ServiceType) (len=8) "NodePort",
  DeprecatedPublicIPs: ([]string) <nil>,
  SessionAffinity: (api.ServiceAffinity) (len=4) "None"
 },
 Status: (api.ServiceStatus) {
  LoadBalancer: (api.LoadBalancerStatus) {
   Ingress: ([]api.LoadBalancerIngress) <nil>
  }
 }
}
```

Annotation : 

```
(map[string]string) (len=4) {
 (string) (len=13) "external/name": (string) (len=11) "service-one",
 (string) (len=13) "external/port": (string) (len=4) "8080",
 (string) (len=12) "external/uri": (string) (len=13) "/path1/child/",
 (string) (len=13) "external/ytpe": (string) (len=4) "http"
}
```

shoud produce

```
frontend frontend_one
    bind *:8080
    acl has_special_uri path_beg /path1/child/
    use_backend special_server if has_special_uri
    default_backend webfarm

backend webfarm
    balance roundrobin
    cookie SERVERID insert
    option httpchk HEAD /check.txt HTTP/1.0 // TODO how to get the health check?
    option httpclose
    option forwardfor
    server webA 10.254.23.4:80 cookie webA check
    server webB 10.248.23.128:80 cookie webB check

backend special_server
    balance roundrobin
    cookie SERVERID insert
    option httpchk HEAD /check.txt HTTP/1.0 // TODO how to get the health check?
    option httpclose
    option forwardfor
    server webC 10.0.0.1:80 cookie webC check
```
