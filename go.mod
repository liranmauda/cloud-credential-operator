module github.com/openshift/cloud-credential-operator

go 1.16

require (
	cloud.google.com/go v0.65.0
	github.com/Azure/azure-sdk-for-go v31.1.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.1
	github.com/Azure/go-autorest/autorest/adal v0.9.5
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2
	github.com/Azure/go-autorest/autorest/date v0.3.0
	github.com/Azure/go-autorest/autorest/to v0.3.0
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/aws/aws-sdk-go v1.37.14
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.4.4
	github.com/google/uuid v1.1.2
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/openshift/api v0.0.0-20201103184615-27004eede929
	github.com/openshift/build-machinery-go v0.0.0-20210702090207-9c7b89e8633a
	github.com/openshift/library-go v0.0.0-20200911100307-610c6e9e90b8
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/mod v0.3.0
	golang.org/x/oauth2 v0.0.0-20210622215436-a8dc77f794b6
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/api v0.30.0
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a
	google.golang.org/grpc v1.31.0
	gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/code-generator v0.20.1
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.3
)
