package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"

	"github.com/colindev/wshub"

	compute "google.golang.org/api/compute/v1"
	dns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"

	"golang.org/x/oauth2/google"
)

type key int

var (
	client         *http.Client
	computeService *compute.Service
	hub            *wshub.Hub
	env            = struct {
		ctx       context.Context
		Addr      string
		ProjectID string
		Region    string
		Zone      string

		DomainName        string
		DNSManageZoneName string

		OSFamily  string
		OSProject string
		OSVersion string

		BasePath       string
		SelfInternalIP string
	}{
		ctx:       context.Background(),
		ProjectID: "gcetest-156204",
		Region:    "asia-east1",
		Zone:      "asia-east1-a",

		DomainName:        "jan23.me",
		DNSManageZoneName: "test-jan23-me",

		OSFamily:  "centos",
		OSProject: "centos-cloud",
		OSVersion: "centos-7-v20170227",

		SelfInternalIP: "127.0.0.1",
	}
)

func getClient() *http.Client {
	return client
}

func getComputeService() *compute.Service {
	return computeService
}

func getWSHub() *wshub.Hub {
	return hub
}

// ProcessStatus ...
type ProcessStatus struct {
	Active string `json:"active"`
	Items  Items  `json:"items"`
}

// Items ...
type Items map[string]string

func init() {
	log.SetFlags(log.Lshortfile)

	var err error
	if metadata.OnGCE() {
		env.ProjectID, err = metadata.ProjectID()
		if err != nil {
			log.Fatal(err)
		}
		env.SelfInternalIP, err = metadata.InternalIP()
		if err != nil {
			log.Fatal(err)
		}
	}

	basePath, err := filepath.Abs("./")
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(basePath + "/public")
	if err != nil {
		basePath = filepath.Dir(os.Args[0])
	}

	env.BasePath = basePath
	flag.StringVar(&env.Addr, "addr", ":80", "http address")

	client, err = google.DefaultClient(env.ctx, compute.ComputeScope, dns.NdevClouddnsReadwriteScope)
	if err != nil {
		log.Fatal(err)
	}

	computeService, err = compute.New(client)
	if err != nil {
		log.Fatal(err)
	}

	hub = wshub.New()
	go hub.Run()
}

func main() {

	flag.Parse()
	log.Printf("%+v", env)

	// views
	fsServer := http.FileServer(http.Dir(env.BasePath + "/public"))

	tpls := map[string]*template.Template{}
	(func() {
		mainTpl := env.BasePath + "/templates/main.tpl"
		tplDir := env.BasePath + "/templates/contents"
		fs, err := ioutil.ReadDir(tplDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range fs {
			name := f.Name()
			if strings.HasSuffix(name, ".tpl") {
				tpls[strings.TrimSuffix(name, ".tpl")] = template.Must(template.ParseFiles(mainTpl, tplDir+"/"+name))
			}

		}
	})()
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimLeft(r.URL.Path, "/")
		name = strings.TrimSuffix(name, ".html")
		if name == "" {
			name = "index"
		}
		if tpl, ok := tpls[name]; ok {
			tpl.Execute(w, nil)
			return
		}
		fsServer.ServeHTTP(w, r)
	}))
	// ws
	http.Handle("/ws", hub.Handler())
	http.HandleFunc("/ws-broadcast", broadcast)
	// admin apis
	http.HandleFunc("/admin/api/compute/zones", listZones)
	http.HandleFunc("/admin/api/compute/images", listImages)
	http.HandleFunc("/admin/api/compute/instances", listComputeInstances)
	http.HandleFunc("/admin/api/compute/instance", getComputeInstance)
	http.HandleFunc("/admin/api/compute/instances/insert", insertComputeInstance)
	http.HandleFunc("/admin/api/compute/instances/delete", deleteConputeInstance)
	http.HandleFunc("/admin/api/address", getAddress)
	http.HandleFunc("/admin/api/addresses/insert", insertAddress)
	// run server
	log.Println(http.ListenAndServe(env.Addr, nil))
}

func writeRes(w http.ResponseWriter, v interface{}) {

	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func listZones(w http.ResponseWriter, r *http.Request) {

	service := getComputeService()
	if service == nil {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project": env.ProjectID,
	}
	query.Read(r.URL.Query())

	res, err := service.Zones.List(query["project"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, res)
}

func listImages(w http.ResponseWriter, r *http.Request) {

	service := getComputeService()
	if service == nil {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project": env.OSProject,
	}
	query.Read(r.URL.Query())

	res, err := service.Images.List(query["project"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, res)
}

func getComputeInstance(w http.ResponseWriter, r *http.Request) {

	service := getComputeService()
	if service == nil {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project": env.ProjectID,
		"zone":    "",
		"name":    "",
	}
	query.Read(r.URL.Query())

	res, err := service.Instances.Get(query["project"], query["zone"], query["name"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, res)
}

func listComputeInstances(w http.ResponseWriter, r *http.Request) {

	service := getComputeService()
	if service == nil {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project": env.ProjectID,
		"zone":    "",
	}
	query.Read(r.URL.Query())

	res, err := service.Instances.List(query["project"], query["zone"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, res)
}

func insertComputeInstance(w http.ResponseWriter, r *http.Request) {

	service := getComputeService()
	if service == nil {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project":        env.ProjectID,
		"zone":           env.Zone,
		"cpu":            "1",    // vCPU
		"memory":         "1024", // MB
		"network":        "default",
		"subdomain":      "",
		"cloud_image":    env.OSProject,
		"image":          env.OSVersion,
		"name":           "",
		"startup_script": "",
	}
	r.ParseForm()
	ctx := query.Read(r.Form).WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, "service", service)
	ctx = context.WithValue(ctx, "region", zone2region(query["zone"]))
	ctx = context.WithValue(ctx, "address_name", query["subdomain"]+"-"+strings.Replace(env.DomainName, ".", "-", -1))

	machineType := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/machineTypes/custom-%s-%s",
		query["project"],
		query["zone"],
		query["cpu"],
		query["memory"])

	imageURL := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/images/%s",
		query["cloud_image"],
		query["image"])

	network := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", query["project"], query["network"])

	callbackURL := "http://" + env.SelfInternalIP + "/ws-broadcast"
	b, err := ioutil.ReadFile(env.BasePath + "/startup_script." + env.OSFamily + ".temp")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	script := strings.Replace(string(b), "{{.Callback}}", callbackURL, -1)
	script = strings.Replace(script, "{{.StartupScript}}", query["startup_script"], 1)

	instance := &compute.Instance{
		Name:        query["name"],
		Description: "post via golang server",
		MachineType: machineType,
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    "disk-" + query["name"],
					SourceImage: imageURL,
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			&compute.NetworkInterface{
				AccessConfigs: []*compute.AccessConfig{
					&compute.AccessConfig{
						Type: "ONE_TO_ONE_NAT",
						Name: "External NAT",
					},
				},
				Network: network,
			},
		},
		ServiceAccounts: []*compute.ServiceAccount{
			{
				Email: "default",
				Scopes: []string{
					"https://www.googleapis.com/auth/devstorage.read_only",
				},
			},
		},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "startup-script",
					Value: &script,
				},
				{
					Key:   "callback-url",
					Value: &callbackURL,
				},
			},
		},
		Tags: &compute.Tags{
			Items: []string{"http-server", "https-server"},
		},
	}

	b, _ = json.MarshalIndent(instance, "", "  ")
	fmt.Println(string(b))

	op, err := service.Instances.Insert(query["project"], query["zone"], instance).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	c := make(chan context.Context, 1)
	go checkInstance(ctx, cancel, c)
	go func() {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			return
		case ctx := <-c:
			makeStaticIP(ctx, cancel, c)
			insertDNSCName(ctx, cancel, c)
		}
	}()

	writeRes(w, op)
}

func deleteConputeInstance(w http.ResponseWriter, r *http.Request) {

	service := getComputeService()
	if service == nil {
		http.Error(w, "compute service not found", 500)
		return
	}

	// 刪除千萬別弄預設值
	query := Config{
		"project": "",
		"zone":    "",
		"name":    "",
	}
	r.ParseForm()
	ctx := query.Read(r.Form).WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)

	op, err := service.Instances.Delete(query["project"], query["zone"], query["name"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// TODO delete cname
	// TODO delete static address
	go checkInstance(ctx, cancel, nil)

	writeRes(w, op)
}

func broadcast(w http.ResponseWriter, r *http.Request) {

	hub := getWSHub()
	if hub == nil {
		http.Error(w, "wshub not found", 500)
		return
	}

	var status ProcessStatus
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	b, _ := json.MarshalIndent(status, "", "  ")
	log.Println("receive broadcast")
	log.Println(string(b))
	hub.Broadcast(status)
}

func checkInstance(ctx context.Context, cancel context.CancelFunc, c chan context.Context) {

	project, _ := ctx.Value("project").(string)
	zone, _ := ctx.Value("zone").(string)
	name, _ := ctx.Value("name").(string)

	defer log.Println("quit check instance")

	service := getComputeService()
	if service == nil {
		cancel()
		return
	}
	hub = getWSHub()
	if hub == nil {
		cancel()
		return
	}

	active := "compute#instance#" + name
	items := Items{
		"project": project,
		"zone":    zone,
	}

	for {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			return
		default:
		}
		time.Sleep(time.Second)
		inst, err := service.Instances.Get(project, zone, name).Do()
		if googleapi.IsNotModified(err) {
			hub.Broadcast(ProcessStatus{
				Active: active,
				Items: Items{
					"not-modified": err.Error(),
				},
			})
		} else if err != nil {
			hub.Broadcast(ProcessStatus{
				Active: active,
				Items: Items{
					"error": err.Error(),
				},
			})
			cancel()
			return
		}

		items["status"] = inst.Status
		hub.Broadcast(ProcessStatus{
			Active: active,
			Items:  items,
		})

		if inst.Status == "RUNNING" {
			items["network-ip"] = inst.NetworkInterfaces[0].NetworkIP
			items["nat-ip"] = inst.NetworkInterfaces[0].AccessConfigs[0].NatIP

			if c != nil {
				c <- context.WithValue(ctx, "ip", items["nat-ip"])
			}

			hub.Broadcast(ProcessStatus{
				Active: active,
				Items:  items,
			})
			return
		}
	}

}

func getAddress(w http.ResponseWriter, r *http.Request) {

	service := getComputeService()
	if service == nil {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project": env.ProjectID,
		"region":  env.Region,
		"address": "",
	}
	query.Read(r.URL.Query())

	res, err := service.Addresses.Get(query["project"], query["region"], query["address"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeRes(w, res)
}

func insertAddress(w http.ResponseWriter, r *http.Request) {

	service := getComputeService()
	if service == nil {
		http.Error(w, "compute service not found", 500)
		return
	}

	query := Config{
		"project":      env.ProjectID,
		"region":       env.Region,
		"ip":           "",
		"address_name": "",
	}
	r.ParseForm()
	ctx := query.Read(r.Form).WithContext(context.Background())
	ctx = context.WithValue(ctx, "service", service)
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan context.Context, 1)

	go makeStaticIP(ctx, cancel, c)
	select {
	case <-ctx.Done():
		http.Error(w, ctx.Err().Error(), 500)
	case ctx := <-c:
		res, _ := ctx.Value("operation").(*compute.Operation)
		writeRes(w, res)
	}

}

func makeStaticIP(ctx context.Context, cancel context.CancelFunc, c chan context.Context) {

	select {
	case <-ctx.Done():
		log.Println(ctx.Err())

	case ctx := <-c:
		service, ok := ctx.Value("service").(*compute.Service)
		if !ok {
			cancel()
			return
		}
		project, _ := ctx.Value("project").(string)
		region, _ := ctx.Value("region").(string)
		ip, _ := ctx.Value("ip").(string)
		name, _ := ctx.Value("address_name").(string)
		res, err := service.Addresses.Insert(project, region, &compute.Address{
			Address: ip,
			Name:    name,
		}).Do()

		if err != nil {
			log.Println(err)
			cancel()
			return
		}

		c <- context.WithValue(ctx, "operation", res)
	}

}

func insertDNSCName(ctx context.Context, cancel context.CancelFunc, c chan context.Context) {

	client := getClient()
	if client == nil {
		cancel()
		log.Println("not found client")
		return
	}

	service, err := dns.New(client)
	if err != nil {
		log.Println(err)
		cancel()
		return
	}

	name, _ := ctx.Value("name").(string)
	ip, _ := ctx.Value("ip").(string)
	fullDomain := name + "." + env.DomainName
	project, _ := ctx.Value("project").(string)
	manageZone := env.DNSManageZoneName
	change := &dns.Change{
		Additions: []*dns.ResourceRecordSet{
			{
				Kind: "dns#resourceRecordSet",
				Name: fullDomain,
				Rrdatas: []string{
					ip,
				},
				Type: "CNAME",
				Ttl:  300,
			},
		},
	}

	res, err := service.Changes.Create(project, manageZone, change).Do()

	log.Println(res)
	log.Println(err)
}

func zone2region(zone string) string {
	s := strings.Split(zone, "-")
	s = s[:len(s)-1]
	return strings.Join(s, "-")
}
