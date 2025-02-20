module github.com/ajeetraina/kubearchinspect-docker-extension/backend

go 1.19

replace github.com/labstack/gommon => github.com/labstack/gommon v0.4.0

require (
	github.com/labstack/echo/v4 v4.10.2
	github.com/sirupsen/logrus v1.9.0
	k8s.io/api v0.27.2
	k8s.io/apimachinery v0.27.2
	k8s.io/client-go v0.27.2
)