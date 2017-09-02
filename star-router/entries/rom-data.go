package entries

var romData = map[string]string{
	"aws-creds/":                                "",
	"aws-creds/access_key_id":                   "REDACTED",
	"aws-creds/region":                          "us-west-2",
	"aws-creds/secret_access_key":               "REDACTED",
	"aws-github-queue/":                         "",
	"aws-github-queue/queue-url":                "https://sqs.us-west-2.amazonaws.com/REDACTED/sd-github-inbound",
	"init-script":                               "echo \"Starting up...\"\ninvoke /rom/drv/aws/clone /boot/cfg/aws-creds /n/aws\ninvoke /rom/bin/ray-ssh /rom/bin/ray /n/ray-ssh\ninvoke /n/aws/sqs/receive-message /boot/cfg/aws-github-queue /tmp/message\n: cat /tmp/message/body\necho \"Done booting service\"",
	"services/aws-client/":                      "",
	"services/aws-client/input-path":            "/boot/cfg/aws-creds",
	"services/aws-client/mount-path":            "/n/aws",
	"services/aws-client/path":                  "/rom/drv/aws",
	"services/aws-ns/":                          "",
	"services/aws-ns/input-path":                "/boot/cfg/aws-creds",
	"services/aws-ns/mount-path":                "/n/aws-ns",
	"services/aws-ns/path":                      "/rom/drv/aws-ns",
	"services/export/":                          "",
	"services/export/input-path":                "/n",
	"services/export/mount-path":                "/n/nsexport",
	"services/export/path":                      "/rom/drv/nsexport",
	"services/httpd/":                           "",
	"services/httpd/input-path":                 "/",
	"services/httpd/mount-path":                 "/n/httpd",
	"services/httpd/path":                       "/rom/bin/httpd",
	"services/hue-client/":                      "",
	"services/hue-client/input-path":            "/boot/cfg/hue-bridge",
	"services/hue-client/mount-path":            "/n/hue",
	"services/hue-client/path":                  "/rom/drv/hue/init-bridge",
	"services/kubernetes-apt/":                  "",
	"services/kubernetes-apt/input/":            "",
	"services/kubernetes-apt/input/config-path": "/home/daniel/.kube/config",
	"services/kubernetes-apt/mount-path":        "/n/kube-apt",
	"services/kubernetes-apt/path":              "/rom/drv/kubernetes",
	"services/os-fs/":                           "",
	"services/os-fs/input/":                     "",
	"services/os-fs/input/host-path":            "../web-frontend/",
	"services/os-fs/mount-path":                 "/n/osfs",
	"services/os-fs/path":                       "/rom/drv/host-os/fs",
	"services/ray-ssh/":                         "",
	"services/ray-ssh/input-path":               "/rom/bin/ray",
	"services/ray-ssh/mount-path":               "/n/ray-ssh",
	"services/ray-ssh/path":                     "/rom/bin/ray-ssh",
	"services/redis-ns/":                        "",
	"services/redis-ns/input/":                  "",
	"services/redis-ns/input/address":           "apt:31500",
	"services/redis-ns/mount-path":              "/n/redis-ns",
	"services/redis-ns/path":                    "/rom/drv/redis-ns",
	"system/":                                   "",
	"system/name":                               "Protostar",
	"system/owner":                              "dan@danopia.net",
}
