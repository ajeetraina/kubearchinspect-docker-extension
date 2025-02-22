module github.com/ajeetraina/kubearchinspect-docker-extension/backend

go 1.19

replace (
    k8s.io/api => k8s.io/api v0.27.2
    k8s.io/apimachinery => k8s.io/apimachinery v0.27.2
    k8s.io/client-go => k8s.io/client-go v0.27.2
)

require (
    github.com/labstack/echo/v4 v4.10.2
    github.com/sirupsen/logrus v1.9.0
    k8s.io/apimachinery v0.27.2
    k8s.io/client-go v0.27.2
)

require (
    github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
    github.com/labstack/gommon v0.4.0 // indirect
    github.com/mattn/go-colorable v0.1.13 // indirect
    github.com/mattn/go-isatty v0.0.17 // indirect
    github.com/valyala/bytebufferpool v1.0.0 // indirect
    github.com/valyala/fasttemplate v1.2.2 // indirect
    golang.org/x/crypto v0.6.0 // indirect
    golang.org/x/net v0.7.0 // indirect
    golang.org/x/sys v0.5.0 // indirect
    golang.org/x/text v0.7.0 // indirect
    golang.org/x/time v0.3.0 // indirect
)