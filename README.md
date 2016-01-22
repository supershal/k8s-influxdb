# Deployment of influxdb cluster in kubernetes

	The main goals of the project:
	 - create influxdb cluster in kubernetes.
	 - configure influxdb cluster.
	 - Stress Test influxdb cluster.

## Prerequisites
- [docker-machine] (https://docs.docker.com/engine/installation/mac/)
- [kubernetes with vagrant] (http://kubernetes.io/v1.1/docs/getting-started-guides/vagrant.html)

## Deploy influxdb cluster to local kubernetes cluster
- Create "infra" namespace
	
	```
		$ kubectl create -f ./k8s/infra-namespace.yaml
	```
- Set default namespace to "infra"
	
	```
		$ export CONTEXT=$(kubectl config view | grep current-context | awk '{print $2}')
		$ kubectl config set-context $(CONTEXT) --namespace=<insert-namespace-name-here>
	``` 
- Deploy influxdb service and cluster
	- A prebuilt docker image is deployed at: https://hub.docker.com/r/supershal/influxdb/
	1. deploy influxdb service
	```
		$ kubectl create -f ./k8s/influxdb-svc.yaml
	```
	2. deploy influxdb raft (jury) node
	```
		$ kubectl create -f ./k8s/influxdb-raft-rc.yaml
	```
	3. scale raft node to 3. no more than 3 allowed.
	```
		$ kubectl scale --replicas=3 rc/influxdb-raft-rc
	```
	4. deploy influxdb data nodes
	```
		$ kubectl create -f ./k8s/influxdb-peers-rc.yaml
	```
	5. scale data node to any number. we set it up to 4
	```
		$ kubectl scale --replicas=4 rc/influxdb-peers-rc
	```
- Deploy grafana service
	1. deploy grafana service
	```
		$ kubectl create -f ./k8s/influxdb-grafana-svc.yaml
	```
	2. deploy grafana node
	```
		$ kubectl create -f ./k8s/influxdb-grafana-rc.yaml
	```

## Verify cluster setup

- ssh to k8s minion
	```
		$ vagrant ssh minion-1
		OR
		$ ssh vagrant@minion-1 # make sure you have set up vagrant config in your ./ssh/config. use vagrant ssh-config command.
	```
- locate one of the influxdb container and exec
	```
		$ sudo docker ps
		$ sudo docker exec -it /bin/bash <container_id> /bin/bash
	```
- launch influxdb cli and check cluster members
	```
		$ influx
		$ show servers
	```
	It should give all members of the cluser something like.
	```
		> show servers
			id	cluster_addr		raft	raft-leader
			1	10.246.98.49:8088	true	true
			2	10.246.98.50:8088	true	false
			3	10.246.98.51:8088	true	false
			4	10.246.98.53:8088	false	false
			5	10.246.98.52:8088	false	false
			6	10.246.98.54:8088	false	false
	```
## Terminate cluster

- delete influxdb service
	```
	 $ kubectl delete svc influxdb-svc
	``` 
	
- delete influxdb rc
	```
 	 $ kubectl delete rc -l app=influxdb
	```
	
- delete grafana service
	```
	 $ kubectl delete svc influxdb-grafana-svc
	``` 
	
- delete grafana rc
	```
 	 $ kubectl delete rc influxdb-grafana-rc
	```

## Build influxdb docker image

If you are not making any code changes and/or building docker image for influxdb, you can skip this section.

1. Install GO latest version and set up go workspace. [Instructions]
2. Install godep
	
	```
		$ go get -u github.com/tools/godep
	```
3. Restore dependency
	
	```
		$ godep restore
	```
4. Install and test influxdbconfig program locally.
    
     - Build go binary for testing.

      ```
		$ go build -o influxdblocal ./influxdb/main.go
      ```
    - Setup kubectl proxy to proxy influxdb api server
      
      ```
      	$ kubectl proxy --port=9090 &
      ```
    - test locally. 
      
      ```
      	$ LOCAL_PROXY="http://localhost:9090" INFLUXDB_POD_SELECTORS="app=influxdb" NAMESPACE="infra" ./influxdblocal test
      ```
      It will spit out the the influxdb cluster config parameters to console.
4. build influxdbconfig executable from the go program and create docker image for influxdbconfig + influxdb

	- Create docker-machine vm and set docker daemon

	 ```
	 	$ docker-machine create --driver=virtualbox default
	 	$ eval "$(docker-machine env default)"
	 ```

	- build the image

	 ```
		$ ./influxdb/build.sh
	 ```
	This will create image in your docker-machine VM.
5. Upload image to docker hub or directy to k8s-minion.
  - (Faster) Copy image directly to k8s minion. Use this method if you frequently building the image.
    * One time setup for ssh to minion.
    	Go to your kubernetes installation and locate Vagrantfile.
  	
  		```
  			$ vagrant ssh-config
  		```
  		Copy output of above command to your ~/.ssh/config file.
    * copy the image from docker-machine VM to K8S minion.
    	
    	```
    		$ docker save supershal/influxdb:stresstest | ssh vagrant@minion-1 sudo docker load
    	```
  - (Slower) push to docker hub. 
  	
  	 ```
  	 	$ docker push supershal/influxdb:stresstest
  	 ```

## Deploy influxdb cluster to Aws kubernetes cluster
 In progress

## Test cases
 In progress

## Stress Test cases
 In progress

## Useful links

- influxdb clusering:
https://docs.influxdata.com/influxdb/v0.9/guides/clustering/

- Access k8s cluster using apis:
https://github.com/kubernetes/kubernetes/blob/master/docs/user-guide/accessing-the-cluster.md#accessing-the-cluster-api

- K8s: developer guide: 
https://github.com/kubernetes/kubernetes/blob/master/docs/devel/README.md

- setup developer environment for k8s:
https://github.com/kubernetes/kubernetes/blob/master/docs/devel/development.md

- swagger doc for k8s:
http://kubernetes.io/third_party/swagger-ui/

- client library for k8s:
https://github.com/kubernetes/kubernetes/blob/master/docs/devel/client-libraries.md

- browse Golang doc and sourcecode for k8s (or any golang project)
   * install Dash https://kapeli.com/dash 
   * load golang docs for any go project. checkout docs for Dash
