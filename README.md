# Deployment of influxdb cluster in kubernetes

	The main goals of the project:
	 - create influxdb cluster in kubernetes.
	 - configure influxdb cluster.
	 - Stress Test influxdb cluster.
	 - Tune influxdb configs for Apigee requirements.

## Prerequisites
- [docker-machine]
- [kubernetes]

## Deploy influxdb cluster to local kubernetes cluster 

## Build influxdb docker image
If you are not making any code changes and/or building docker image for influxdb, you can skip this section.
1. Install GO latest version and set up go workspace. [Instructions]
2. Install godep
	```
		go get -u github.com/tools/godep
	```
3. Restore dependency
	```
		godep restore
	```
4. Install and test influxdbconfig program locally.
	- Build go binary for testing.
	  ```
		go build -o influxdblocal ./influxdb/main.go
      ```
    - Setup kubectl proxy to proxy influxdb api server
      ```
      	kubectl proxy --port=9090 &
      ```
    - test locally. 
      ```
      	LOCAL_PROXY="http://localhost:9090" INFLUXDB_POD_SELECTORS="app=influxdb" NAMESPACE="infra" influxdblocal test
      ```
      It will spit out the the influxdb cluster config parameters to console.
4. build influxdbconfig executable from the go program and create docker image for influxdbconfig + influxdb
	- Create docker-machine vm and set docker daemon
	 ```
	 	docker-machine create --driver=virtualbox default
	 	eval "$(docker-machine env default)"
	 ```
	- build the image
	 ```
		./build.sh
	 ```
	This will create image in your docker-machine VM.
5. Upload image to docker hub or directy to k8s-minion.
  - (Slower) push to docker hub. 
  	 ```
  	 	docker push spatel/influxdb:stresstest
  	 ```
  - (Faster) Copy image directly to k8s minion. Use this method if you frequently building the image.
    * One time setup for ssh to minion.
    	Go to your kubernetes installation and locate Vagrantfile.
  		```
  			vagrant ssh-config
  		```
  		Copy output of above command to your ~/.ssh/config file.
    * copy the image from docker-machine VM to K8S minion.
    	```
    		docker save spatel/influxdb:stresstest | ssh vagrant@minion-1 sudo docker load
    	```
## Deploy influxdb cluster to Aws kubernetes cluster

## Test cases

## Stress Test cases

## Useful links

build and upload local images to kubernates minion
- docker-machine create --driver virtualbox default  //create machine
- eval "$(docker-machine env default)" // connect to docker
- docker build -t spatel/influxdb:0.1 . //build image
- go to kubernates master folder
- vagrant ssh-config 
- copy the output to ~/.ssh/config
- test connection by ssh vagrant@minion-1 sudo docker ps
- docker save spatel/influxdb:0.2 | ssh vagrant@minion-1 sudo docker load

### execute kubernates commands to create influxdb.
- kubectl create -f influxdb-service.yaml
- kubectl create -f influxdb-rc.yaml



Access k8s cluster using apis:
https://github.com/kubernetes/kubernetes/blob/master/docs/user-guide/accessing-the-cluster.md#accessing-the-cluster-api

K8s: developer guide. 
https://github.com/kubernetes/kubernetes/blob/master/docs/devel/README.md

setup developer environment:
https://github.com/kubernetes/kubernetes/blob/master/docs/devel/development.md

swagger doc:
http://kubernetes.io/third_party/swagger-ui/

client library:
https://github.com/kubernetes/kubernetes/blob/master/docs/devel/client-libraries.md



# run program locally
- setting proxy
  kubectl proxy --port=8080 &
- set env variables.
  export 


