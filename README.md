
[![Join the chat at https://gitter.im/tleyden/couchbase-cluster-go](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/tleyden/couchbase-cluster-go?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

This is a Go library that helps initialize and manage a Couchbase Server cluster running under CoreOS.


## Instructions 

* [Couchbase Server CoreOS README](https://github.com/couchbaselabs/couchbase-server-coreos/)

## Power Tips

### Running on the latest code

Since the docker image can be out of date, and rebuilding it can be time consuming, there is a way to force couchbase-cluster-go to run the latest version of the code:

```
$ etcdctl set /couchbase.com/enable-code-refresh true
$ sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper couchbase-fleet launch-cbs \
  --version 3.0.1 \
  --num-nodes 3 \
  --userpass "user:passw0rd" 
$ sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper sync-gw-cluster launch-sgw \
  --num-nodes=1 \
  --config-url=http://git.io/b9PK \
  --create-bucket todos \
  --create-bucket-size 512 \ 
  --create-bucket-replicas 1
```

### Running Sync Gateway behind an Nginx proxy

You need to pass another parameter: `--launch-nginx` when launching Sync Gateway, and you also need to be running on the latest code.

The easiest way is to replace the third command in **Running on the latest code** with:

```
$ sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper sync-gw-cluster launch-sgw --launch-nginx --num-nodes=1 --config-url=http://git.io/b9PK --create-bucket todos --create-bucket-size 512 --create-bucket-replicas 1
```

**Verify Internal**

After the Sync Gateway instance(s) launch, find the IP of the nginx node by:

```
$ fleetctl list-units
UNIT				MACHINE				ACTIVE	SUB
...
nginx.service			5c7662f4.../10.136.111.112	active	running
...
```

From one of the machines in the coreos cluster, try issuing a request to port 80 of that ip:

```
$ curl 10.136.111.112
{"couchdb":"Welcome","vendor":{"name":"Couchbase Sync Gateway","version":1},"version":"Couchbase Sync Gateway/master(a47a17f)"}
```

**Verify External**

Finally, you can find the public ip and then pasting the ip into your web browser on your workstation.  If it doesn't work, you may need to update your AWS Security Group to allow access to port 80 from any ip address.

You might also want to change the default Security Group to change port 4984 to only be accessible from within the CloudFormation group (as opposed to accessible from any address). 

### Sync Gateway -> Couchbase Server service discovery

There is a mechanism that will rewrite the Sync Gateway config provided before launching the Sync Gateway.  To leverage this, simply modify your Sync Gateway config so that the `server` field contains `http://{{ .COUCHBASE_SERVER_IP }}:8091`.  

A live Couchbase Server node will be discovered via etcd and the value in the Sync Gateway config will be replaced with that node's ip address.

[Complete Sync Gateway Config example](https://gist.github.com/tleyden/ca063725e6158eca4093)

### Destroying the cluster

The following commands will stop and destroy all units (Couchbase Server, Sync Gateway, and otherwise)

```
$ sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper couchbase-fleet stop --all-units
$ sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper couchbase-fleet destroy --all-units
```

This command will delete all persistent data in the `/opt/var/couchbase` directory across all machines on the cluster.

**WARNING - this will destroy all of your data stored in Couchbase**

```
fleetctl list-machines | grep -v MACHINE | awk '{print $2}' | xargs -I{} ssh {} 'sudo rm -rf /opt/couchbase/var/'
```

### Workaround fleetctl start issues

Units will frequently come up as failed due to [fleet issue 1149](https://github.com/coreos/fleet/issues/1149).  To workaround couchbase_sidekick@1.service coming up as failed, do the following:

```
$ fleetctl cat couchbase_sidekick@1.service > couchbase_sidekick@1.service
$ fleetctl stop couchbase_sidekick@1.service
$ fleetctl destroy couchbase_sidekick@1.service
$ fleetctl start couchbase_sidekick@1.service
```

The last command will use the unit file saved in the first step.


## Issue Tracker

Please file issues to the [couchbase-server-coreos](https://github.com/couchbaselabs/couchbase-server-coreos) repo.  
## Related Work

* [couchbase-array](https://github.com/andrewwebber/couchbase-array) -- community user contributed alternative solution
