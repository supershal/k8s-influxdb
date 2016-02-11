package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	client "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

var (
	//os envs for testing locally.
	localProxy = os.Getenv("LOCAL_PROXY")
	// os envs for getting influx db pods
	influxSelectors   = os.Getenv("INFLUXDB_POD_SELECTORS") // app=influxdb,type=raft
	namespace         = os.Getenv("NAMESPACE")              // infra
	influxClusterPort = "8091"
	envVarFile        = "/etc/default/influxdb"
)

func main() {
	if namespace == "" {
		log.Fatalf("NAMESPACE env not set")
		os.Exit(2)
	}
	if influxSelectors == "" {
		log.Fatalf("INFLUXDB_POD_SELECTOR env not set")
		os.Exit(2)
	}
	Execute()
}

// This represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "influxdb",
	Short: " set cluster config for influxdb",
	Long:  `Get pods from k8s apis for influxdb and create config in /etc/default/influxdb to join the pod to the cluster`,
}

func Execute() {
	rootCmd.AddCommand(joinCmd)
	rootCmd.AddCommand(testCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
		os.Exit(-1)
	}
}

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join the influxdb cluster",
	Long:  "Generate dynamic list of peer influxdb pods using k8s api and write peers in the /etc/default/influxdb",
	RunE:  runJoinCluster,
}

func runJoinCluster(cmd *cobra.Command, args []string) error {
	cli, err := client.NewInCluster()
	if err != nil {
		return fmt.Errorf("unable to connect k8s api server: %v", err)
	}

	labelSelector, err := labels.Parse(influxSelectors)
	if err != nil {
		return fmt.Errorf("unable to parse labels: %v", err)

	}
	fieldSelector := fields.Everything()
	podIPs, err := podIps(cli, labelSelector, fieldSelector)

	if err != nil {
		return err
	}

	hostIP, err := externalIP()
	if err != nil {
		return err
	}
	peers := influxdbPeers(hostIP, podIPs)
	iOpts := influxdOpts(hostIP, peers)

	if err := ioutil.WriteFile(envVarFile, []byte(iOpts), 0644); err != nil {
		return err
	}
	return nil
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Spit out content of /etc/default/influxdb",
	Long: `Please set k8s proxy using: "kubectl proxy --port=8989 &".
	Generate dynamic list of peer influxdb pods using k8s api and write peers in the /etc/default/influxdb`,
	RunE: runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	if localProxy == "" {
		return fmt.Errorf("please set env variable LOCAL_PROXY. ex: LOCAL_PROXY=\"http://localhost:8080\". Setup proxy as \"kubectl proxy --port 8080 &\" ")
	}

	config := &client.Config{
		Host:     localProxy,
		Insecure: true,
	}

	cli, err := client.New(config)
	if err != nil {
		return fmt.Errorf("unable to connect k8s api server: %v", err)
	}

	labelSelector, err := labels.Parse(influxSelectors)
	if err != nil {
		return fmt.Errorf("unable to parse labels: %v", err)

	}
	fieldSelector := fields.Everything()
	if podIPs, err := podIps(cli, labelSelector, fieldSelector); err != nil {
		return err
	} else {
		hostIP, _ := externalIP()
		peers := influxdbPeers(hostIP, podIPs)
		iOpts := influxdOpts(hostIP, peers)

		fmt.Println("Content of /etc/default/influxdb : ", iOpts)
	}
	return nil
}

//TODO: check running pod ips.
func podIps(cli *client.Client, lblSel labels.Selector, fieldSel fields.Selector) ([]string, error) {
	podOptions := &api.ListOptions{
		LabelSelector: lblSel,
		FieldSelector: fieldSel,
	}

	pods, err := cli.Pods(namespace).List(*podOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve pods: %v", err)
	}
	ipList := make([]string, 0.0)
	for _, pod := range pods.Items {
		if pod.Status.Phase != api.PodRunning {
			continue
		}
		ipList = append(ipList, pod.Status.PodIP)
	}
	return ipList, nil
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

func influxdbPeers(hostIP string, podIPs []string) []string {
	var peers []string
	for _, peer := range podIPs {
		if peer == hostIP {
			continue
		}
		peers = append(peers, peer)
	}
	return peers
}

// INFLUXD_OPTS="-join hostname_1:port_1,hostname_2:port_2"
func influxdOpts(_ string, peers []string) string {
	if len(peers) == 0 {
		return ""
	}
	var acc string

	// max 3 peers allowed.
	if len(peers) > 3 {
		peers = peers[:3]
	}

	for _, peer := range peers {
		acc = acc + peer + ":" + influxClusterPort + ","
	}
	acc = acc[:len(acc)-1]
	joinPeersOpt := "-join " + acc
	influxdOpt := joinPeersOpt
	return "INFLUXD_OPTS=\"" + influxdOpt + "\""
}

// // INFLUXD_OPTS="-join hostname_1:port_1,hostname_2:port_2 -hostname hostIP:port"
// ENABle below function for influxdb < 0.10
// func influxdOpts(hostIP string, peers []string) string {
// 	hostvar := "-hostname " + hostIP + ":" + influxClusterPort
// 	if len(peers) == 0 {
// 		return "INFLUXD_OPTS=\"" + hostvar + "\""
// 	}
// 	var acc string

// 	// max 3 peers allowed.
// 	if len(peers) > 3 {
// 		peers = peers[:3]
// 	}

// 	for _, peer := range peers {
// 		acc = acc + peer + ":" + influxClusterPort + ","
// 	}
// 	acc = acc[:len(acc)-1]
// 	joinPeersOpt := "-join " + acc
// 	influxdOpt := joinPeersOpt + " " + hostvar
// 	return "INFLUXD_OPTS=\"" + influxdOpt + "\""
// }
