# Help

## <a name="9815029643"></a>Installing and setting up RuntimeRadar

This section contains instructions on how to install and set up Runtime Radar.

### Hardware and software requirements

Hardware requirements are based on load testing results. Tests were performed using a server with the default number of pods specified in the Helm chart. The requirements can change depending on the following factors:
* Number of generated runtime events
* Number of generated events from runtime sources
* Number of namespaces and pods

Each pod with containers of third-party applications inevitably but unpredictably increases the number of events. This leads to increased hardware requirements.

**Hardware requirements for the deployment server**

Hardware requirements for the Runtime Radar deployment server are provided in the table below.

<table><caption>Hardware requirements for the Runtime Radar deployment server</caption><colgroup><col style="width: 50.0%;"/><col style="width: 50.0%;"/></colgroup><thead><tr><th align="left">

Parameter
</th><th align="left">

Minimum requirements
</th></tr></thead><tbody><tr><td align="left">

CPU
</td><td align="left">

4 cores
</td></tr><tr><td align="left">

RAM
</td><td align="left">

8 GB
</td></tr><tr><td align="left">

HDD
</td><td align="left">

20 GB
</td></tr></tbody></table>


You must establish connection to the Kubernetes cluster API and to an artifactory with the Runtime Radar images.

**Hardware requirements for a server with the central Kubernetes cluster**

Hardware requirements for a server with the central Kubernetes cluster are in the table below.

> **Warning.** The specified requirements are necessary for Runtime Radar to run and do not take into account the load already present on the server with the central Kubernetes cluster.

<table><caption>Hardware requirements for a server with the central Kubernetes cluster</caption><colgroup><col style="width: 33.3%;"/><col style="width: 33.3%;"/><col style="width: 33.3%;"/></colgroup><thead><tr><th align="left">

Parameter
</th><th align="left">

Minimum requirements
</th><th align="left">

Optimal configuration
</th></tr></thead><tbody><tr><td align="left">

CPU
</td><td align="left">

8 cores
</td><td align="left">

10 cores
</td></tr><tr><td align="left">

RAM
</td><td align="left">

10 GB
</td><td align="left">

16 GB
</td></tr><tr><td align="left">

HDD
</td><td align="left">

40 GB
</td><td align="left">

100–150 GB
</td></tr></tbody></table>


If you use PostgreSQL, RabbitMQ, and ClickHouse from the Runtime Radar installation distribution package, allocate additional free resources in the central Kubernetes cluster as outlined in the table below.

<table><caption>Hardware requirements for additional resources in the central Kubernetes cluster when PostgreSQL, RabbitMQ, and ClickHouse from the Runtime Radar installation distribution package are used</caption><colgroup><col style="width: 50.0%;"/><col style="width: 50.0%;"/></colgroup><thead><tr><th align="left">

Parameter
</th><th align="left">

Minimum requirements
</th></tr></thead><tbody><tr><td align="left">

CPU
</td><td align="left">

8 cores
</td></tr><tr><td align="left">

RAM
</td><td align="left">

16 GB
</td></tr></tbody></table>


The minimum disk size is calculated using different formulas depending on the stored data.

The average size of a Tetragon event is 13 KB. You can calculate the minimum HDD size for a ClickHouse server depending on the average number of events per second within the calculated period (R) and the number of days to store events for (d) using the following formula: HDD = 13 × R × 86 400 × d.

The calculateed HDD size is in KB. For the size in GB, divide the value by 10<sup>6</sup>.

The minimum HDD size for a PostgreSQL server depends on the following:
* Size of an image scan event (depends on the number of components and vulnerabilities), s. The average event size is 500 KB. For images with a large number of components, we recommend that you increase this value by 25-30%.
* Number of images to scan during a scheduled scan, I.
* Number of scans within the retention period, N<sub>d</sub>. For example, if one scheduled scan is performed per week, 52 scans will be performed per year and 4 scans—per month. If a scan is performed twice a day, 62 scans will be performed per month.

You can calculate the minimum HDD disk size for a PostgreSQL server using the following formula: HDD = s × I × N<sub>d</sub>.

The calculateed HDD size is in KB. For the size in GB, divide the value by 10<sup>6</sup>.

The minimum HDD size for a server that stores admission controller check events and configuration scans depends on the following:
* Kubernetes API server RPS to the admission controller module, N<sub>r</sub>. The RPS can be determined using [Grafana](#9502135051).
* Scan event size, S<sub>s</sub>. As a rule, a scan event contains information about one component (except IaC) so the average event size is 50–60 KB.
* Event retention period, d.

You can calculate the minimum HDD disk size for the server to store admission controller check events and to scan configurations using the following formula: HDD = N<sub>r</sub> × S<sub>s</sub> × 86 400 × d.

The calculateed HDD size is in KB. For the size in GB, divide the value by 10<sup>6</sup>.

If you use PostgreSQL, RabbitMQ, ClickHouse, or Redis obtained from somewhere other than the Runtime Radar installation distribution package, ensure that the cluster is connected to the servers of the corresponding DBMSs.

**Hardware requirements for servers with child Kubernetes clusters**

The hardware requirements for servers with child Kubernetes clusters are identical to the requirements for a server with the central cluster except for the network access requirements. Servers with child Kubernetes clusters require network access to a server with the central cluster.

**Hardware requirements for a server with the Runtime Radar image artifactory **

A server with the Runtime Radar image artifactory requires at least 3 GB of free disk space.

**Hardware requirements for a DBMS server**

If additional parameters are not specified, the DBMS is deployed in the Kubernetes cluster. When deploying Runtime Radar, you can specify settings for connecting to existing DBMSs. On the servers with external PostgreSQL and ClickHouse DBMSs, allow SSL connection. The servers must meet the requirements provided in the table below.

<table><caption>Hardware requirements for a DBMS server</caption><colgroup><col style="width: 38.6%;"/><col style="width: 61.3%;"/></colgroup><thead><tr><th align="left">

Parameter
</th><th align="left">

Minimum value
</th></tr></thead><tbody><tr><td align="left">

CPU
</td><td align="left">

8 cores
</td></tr><tr><td align="left">

RAM
</td><td align="left">

16 GB
</td></tr><tr><td align="left">

HDD
</td><td align="left">

500 GB
</td></tr><tr><td colspan="2" align="center">

SSD
</td></tr><tr><td align="left">

Bandwidth
</td><td align="left">

Read 313 MB/s, write 156 MB/s
</td></tr><tr><td align="left">

IOPS
</td><td align="left">

10,000 read operations per second, 5,000 write operations per second
</td></tr></tbody></table>


**Software requirements for the deployment server**

On the deployment server, place the `kubeconfig` configuration file for connection to Kubernetes and install the Helm package manager.

> **Warning.** The Helm version must be compatible with the Kubernetes version you use. For compatibility information, see the [Helm](https://helm.sh/docs/) documentation. Use of incompatible versions may lead to unexpected consequences.

**Software requirements for the central Kubernetes cluster and child Kubernetes clusters**

Use Kubernetes version 1.20 or later and the following DBMSs:
* Redis version 7.0.11 or later
* PostgreSQL version 14 or later
* ClickHouse version 20 or later.

If you use the PostgreSQL, ClickHouse, and Redis DBMSs and the RabbitMQ server from the Runtime Radar installation distribution package, then, during deployment, specify storage classes for them.

You must deploy external databases and RabbitMQ version 3.12.6 beforehand and specify their connection parameters when installing Runtime Radar.

**Software requirements for the OS kernel of the Kubernetes cluster nodes**

Use an OS whose core meets the following requirements:
* OS kernel version supports Tetragon.

   ***Note.** It is recommended that you use an OS with kernel version 5.4 or later.*
* BTF feature is enabled in the kernel.
* eBPF and cgroup modules are enabled in the kernel. For more information about the modules, see the official Tetragon documentation.
* Linux security modules (LSMs) that control the launch of eBPF programs are disabled in the kernel or run in the allow mode:
   * For the lockdown module, the mode is set to `none` or `integrity`.
   * For the Yama module, the mode is set to `0` or `1`.
   * The AppArmor, SElinux, and Parsec modules are disabled for the Tetragon processes, or accurate access to kernel modules is set up.

### Quick installation using Helm

The Helm chart configuration file with the default settings will be used for installation. If you need to consider the specifics of the existing infrastructure, you can manually fill in the Helm chart configuration file prior to installation. All of the available settings are described in the `README.md` file.

To install Runtime Radar using Helm,

1. Run the following command:

   ```bash
   helm install runtime-radar -n runtime-radar --create-namespace oci://ghcr.io/runtime-radar/runtime-radar:v0.2.0 \
     --set-string 'global.ownCsUrl=https://<your domain address>:32000' \
     --set-string 'global.keys.publicAccessTokenSalt=INIT-DO-NOT-USE' \
     --set-string 'global.keys.encryption=INIT-DO-NOT-USE' \
     --set-string 'auth-center.administrator.username=admin' \
     --set-string 'auth-center.administrator.password=Password' \
     --set-string 'history-api.retentionInterval=8760h' \
     --set-string 'postgresql.auth.username=admin' \
     --set-string 'postgresql.auth.password=Password' \
     --set-string 'postgresql.auth.database=rr_quickstart' \
     --set 'postgresql.persistence.enabled=false' \
     --set-string 'redis.auth.username=admin' \
     --set-string 'redis.auth.password=Password' \
     --set 'redis.persistence.enabled=false' \
     --set-string 'rabbitmq.auth.username=admin' \
     --set-string 'rabbitmq.auth.password=Password' \
     --set 'rabbitmq.persistence.enabled=false' \
     --set 'clickhouse.deploy=true' \
     --set-string 'clickhouse.auth.username=admin' \
     --set-string 'clickhouse.auth.password=Password' \
     --set-string 'clickhouse.auth.database=rr_quickstart' \
     --set 'clickhouse.persistence.enabled=false' \
     --set-string 'reverse-proxy.service.type=NodePort' \
     --set-string 'reverse-proxy.service.nodePorts.http=32000'
   ```

   ***Note.** In the command example, the username is `admin` and the password is `Password`. You can specify other values and later use them to connect to the Runtime Radar web interface.*

   ***Note.** In the command example, access to the web interface is configured using the NodePort service on port 32000. You can use the Ingress controller instead or change the port. To do this, you must specify the [corresponding settings](#9815029643). You can also change [other settings](#9815029643) in the installation command.*

Now you can start setting up the runtime event monitoring.

### Setting up the Runtime Radar web interface connection

You can set up access to the Runtime Radar web interface using the following methods:
* Using port forwarding.
* Using the Ingress controller. Runtime Radar supports two types of Ingress controllers: nginx and HAProxy. To use other Ingress controllers, you need to configure additional settings manually.
* Using the NodePort service.

**Setting up access to the web interface using port forwarding**

To set up access to the web interface using port forwarding,

1. Run the command:

   ```
   kubectl -n runtime-radar port-forward svc/reverse-proxy 9000:9000
   ```

   Information about the port appears through which the Runtime Radar web interface can be accessed.

**Setting up access to the web interface using the Ingress controller**

To configure access to the web interface using the Ingress controller:

1. Install the Ingress controller beforehand.

1. In the Runtime Radar installation command, specify the name of the installed Ingress controller in the `--ingress-class` parameter.

1. In required, in the Runtime Radar installation command, set the `--deploy-ingress` parameter to `0` and create an object of the Ingress type with custom annotations.

   ***Note.** To set up access to the web interface using the Ingress controller, you can use the certificate that already exists in the infrastructure; to do this, specify the `--ingress-cert` and `--ingress-cert-key` parameters.*

Example of an Ingress manifest with parameters for nginx:

```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: rr-reverse-proxy-ingress
  labels:
    app.kubernetes.io/name: reverse-proxy
    app.kubernetes.io/instance: runtime-radar
  annotations:
    # nginx default configs
    # https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/
    nginx.ingress.kubernetes.io/proxy-body-size: "100m"
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "1800"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "1800"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "1800"
spec:
  rules:
    - http:
      paths:
        - pathType: Prefix
          path: "/"
          backend:
            service:
            name: reverse-proxy
              port:
                number: 9000
```

***Note.** The example above shows the size threshold of a request body and that of the proxyed connection timeout.*

**Setting up access to the web interface using the NodePort service**

To set up access to the web interface using the NodePort service:

1. During Runtime Radar installation, specify `--deploy-nodeport` for the NodePort service.

1. If required, in the `--nodeport` parameter, enter a specific port for access to the NodePort service.

### Configuring monitoring of runtime events

Manifests are used to configure the Runtime Radar modules. The manifest is a configuration file in YAML format.

**Limiting messages in the RabbitMQ queue**

A large number of messages in the RabbitMQ queue may lead to a significant increase in memory usage. You can specify the number of sent messages using the `RABBIT_RUNTIME_EVENTS_QUEUE_PREFETCH_COUNT` setting in the event processor manifest and the `RABBIT_QUEUE_PREFETCH_COUNT` setting in the history api manifest. The default value is 100.

To configure the number of sent messages:

1. In the edit mode, open the event processor module manifest.

1. Set a `RABBIT_RUNTIME_EVENTS_QUEUE_PREFETCH_COUNT` value.

1. Apply the changes and exit the edit mode.

1. In the edit mode, open the history api module manifest.

1. Set a `RABBIT_QUEUE_PREFETCH_COUNT` value.

1. Apply the changes and exit the edit mode.

1. Restart the event processor and history api modules sequentially by running the following command for each of them:

   ```
   kubectl rollout restart deployment/<module service name> -n <namespace where Runtime Radar is installed>
   ```

**Configuring the runtime monitor buffer size**

If the RabbitMQ module is unavailable, using the buffer allows you to continue processing runtime events. You can set the runtime monitor module buffer size by specifying the `TETRAGON_EVENTS_BUFFER` setting in the module manifest and the `CONFIG_UPDATE_INTERVAL` setting in the configuration map (ConfigMap) of the module.

To configure the runtime monitor buffer size:

1. Open the runtime monitor module manifest in edit mode.

1. Set a `TETRAGON_EVENTS_BUFFER` value.

1. Apply the changes and exit the edit mode.

1. Open the `runtime-monitor-config` configuration card (ConfigMap) in edit mode.

1. Set a `CONFIG_UPDATE_INTERVAL` value.

1. Restart the runtime monitor service by running the following command:

   ```
   kubectl rollout restart daemonset/<service name> -n <namespace where Runtime Radar is installed>
   ```

**Changing network ports for the runtime monitor service**

The runtime monitor service operates in the hostNetwork mode, that is, it uses a network namespace of the host. If a host port is already in use, the service will not track runtime events. You can set ports for the service in the Helm chart configuration file `values.yaml` before installing Runtime Radar. You can also set a port for the tetragon container of the runtime monitor service if the default host port is already in use.

To set a network port for the runtime monitor service:

1. Open the `values.yaml` file in edit mode.

1. In the runtime monitor section, specify port values.

   Example:

   ```
   runtime-monitor:
       containerPorts:
           http: 9000
           grpc: 8000
           gops: 7000
   ```

1. If necessary, specify a port for the tetragon container.

   Example:

   ```
   runtime-monitor:
       containerPorts:
           http: 9000
           grpc: 8000
           gops: 7000
       tetragon:
           grpc:
               address: "localhost:54321"
   ```

1. Apply the changes and exit the edit mode.

   ```
   kubectl rollout restart daemonset/<service name> -n <namespace where Runtime Radar is installed>
   ```

## Managing child clusters

You can use Runtime Radar to protect multiple clusters. You must first install Runtime Radar in the central cluster and then connect child clusters to the Runtime Radar central cluster. This allows you to do the following:
* Simplify incident management in a complex infrastructure using a single management point.
* Provide centralized monitoring and response to information security events.
* Ensure the maximum infrastructure coverage.

The central cluster is a cluster to which child clusters are connected and through requests to which they are managed.

The child cluster is a cluster with a deployed Runtime Radar instance and connected to the central cluster.

The "auth API" and "cluster manager" modules are not installed in child clusters. This ensures high performance and reduces network load. The "auth API" module manages accounts used to manage all connected clusters. The "cluster manager" module manages connection and deletion of child clusters.

Runtime Radar instances installed in child clusters are managed in the web interface of Runtime Radar installed in the central cluster.

***Note.** You must provide network connection between the central cluster and the child clusters for requests to be proxied from the central cluster to the child clusters. If a child cluster is unavailable, the central cluster will not be able to get child cluster data and error 404 will be displayed in the web interface.*

You can view all connected child clusters, connect new clusters, or delete them on the **Clusters** page, which opens when you click **Clusters** on the main menu.

***Note.** Which features are available on the page depends on the user role.*

The page displays the following elements:
* **Cluster** list to select a cluster to manage. The selected value will be set on other pages of the web interface.

   ***Note.** You can also select a cluster in the upper-right corner of all pages of the interface except the **Users** page.*
* **Connect** button to connect a new child cluster and get a command to install the cluster.
* Information about the number of clusters that can be connected according to the license and the number of connected clusters.
* Sections with information about connected clusters.

Each section with information about the connected cluster has the following elements:
* Name of the connected cluster.
* Cluster installation status. The cluster can have one of the following installation statuses: **Not installed** and **Installed**.
* Date when the cluster was created in the web interface.
* Cluster status.
* Buttons to change the cluster name and delete the cluster. If the cluster is selected in the Runtime Radar web interface, the deletion button is inactive.

### <a name="9839873547"></a>Connecting a child cluster to Runtime Radar

To connect a child cluster, you must first add the child cluster in the Runtime Radar web interface and get the installation command. Run the received installation command in the child cluster. Then, the cluster will be automatically registered in Runtime Radar.

**Connecting a child cluster**

To connect a child cluster:

1. On the main menu, select **Clusters**.

1. Click **Connect**.

1. Fill in the connection settings in advance.

1. At the last step, click **Get installation command**.

   The page with the installation commands and the button to download the `values.yaml` file with the Helm chart configuration will open. To install Runtime Radar in a child cluster, you can run the installation command or YAML installation command. To install Runtime Radar using the YAML installation command, you must save the `values.yaml` file with the Helm chart configuration.

1. Complete one of the following actions:

   * Copy the installation command.

   * Copy the YAML installation command.

1. If you copied the YAML installation command, click **Download values.yaml**.

1. Click **Finish**.

**Installation in a child cluster**

You can install Runtime Radar in a child cluster using the installation command or the YAML installation command.

To install Runtime Radar in a child cluster using the installation command:

1. Connect to the cluster where to install Runtime Radar.

1. Run the installation command.

   ***Note.** If necessary, you can change the path to the Helm chart in the installation command manually.*

   After Runtime Radar is installed, the child cluster must be registered automatically. After successful registration, on the **Cluster** page in the web interface, the installation status will change to **Installed**. If the status does not change for more than 15 minutes, examine event logs for the cs-manager module.

To install Runtime Radar in a child cluster using the YAML installation command:

1. Connect to the child cluster where to install Runtime Radar.

   ***Note.** If necessary, you can change the path to the Helm chart in the YAML installation command manually.*

1. Go to the directory where the `values.yaml` file was saved.

1. Run the YAML installation command.

   After Runtime Radar is installed, the child cluster must be registered automatically. After successful registration, on the **Cluster** page in the web interface, the installation status will change to **Installed**. If the status does not change for more than 15 minutes, examine event logs for the cs-manager module.

You can manage child clusters in the web interface of the central cluster. If self-signed certificates were generated during Runtime Radar installation in the central cluster, you will not be able to access the data of the child clusters because the connection security cannot be validated. For that reason, before selecting a child cluster, you must add the root certificate to the trusted certificates or follow the child cluster URL by adding the URL to the security exceptions or ignoring the warning about an insecure connection.

### Deleting a child cluster

If Runtime Radar is not yet installed in a child cluster and the child cluster is not registered in Runtime Radar, you can delete the child cluster from the database and the web interface.

If Runtime Radar has already been installed in the child cluster and the child cluster is registered in Runtime Radar, deletion is performed in the following several stages:
1. Getting a deletion command in the Runtime Radar web interface.
1. Running the deletion command in the child cluster.
1. Deregistering the child cluster in Runtime Radar. Performed automatically.

> **Warning.** After you delete a registered cluster, monitoring and response to runtime events will not be available for the cluster.

**Deleting a child cluster that is not registered in Runtime Radar**

To delete a child cluster that is not registered in Runtime Radar:

1. On the main menu, select **Clusters**.

1. Click ![pic](pics/9697536907.svg) next to the cluster that you want to delete.

   Information about the cluster will be deleted from the database and the Runtime Radar web interface.

**Deleting a child cluster registered in Runtime Radar**

To delete a child cluster registered in Runtime Radar:

1. On the main menu, select **Clusters**.

1. Click ![pic](pics/9697536907.svg) next to the cluster that you want to delete.

   ***Note.** The ![pic](pics/9697536907.svg) button is inactive if the cluster is selected in the **Cluster** list in the top right corner of the page.*

1. Copy the resulting deletion command.

1. Connect to a child cluster using the `kubectl` utility or other cluster management utilities.

1. Run the copied deletion command.

   The child cluster will be deregistered in Runtime Radar and deleted from the web interface automatically after the deletion command is executed.

## Managing notification services and templates

You can integrate Runtime Radar into existing infrastructure and container security processes using email notifications, the webhook tool, or syslog protocol.

Email messages notify a narrow range of users (for example, the employees of one department) about information security events.

Webhook notifications allow you to implement integration with services such as Slack, Mattermost, Telegram, or higher-level systems such as SOAR systems (Security Orchestration, Automation, and Response). If Runtime Radar is installed in a network segment isolated from the internet, you can implement webhook notifications through a proxy server by specifying its address during Runtime Radar installation.

You can integrate Runtime Radar into information security monitoring processes through configuring of notifications to SIEM systems or ILMSs.

You can manage notification services and templates in the Runtime Radar web interface on the **Notification services** page, which opens as you select the **Notification services** section on the menu.

***Note.** Which features are available on the page depends on the user role.*

The page displays the **Cluster** list for selecting a cluster and the **New service** button for connecting the notification service. If child clusters are not yet connected to Runtime Radar or they are not yet in the **Installed** status, you can select only the **Central** option.

The **Notification services** page displays all connected services that are used for notifications based on templates, and information about the added notification templates. Notifications are configured [in a response rule](#7481272715).

The connected notification services are grouped by type (for example, email, webhook, or syslog). The name of each service is displayed with the number of notification templates for which it is used. For each service, there is the ![pic](pics/9810272139.svg) button for adding a notification template and the ![pic](pics/9783588875.svg) button for editing or deleting the service.

You can delete a notification service that is no longer in use. All notification templates created for the service will be deleted together with it.

> **Warning.** A notification service cannot be deleted if at least one of its notification templates is used [in an existing response rule](#7481272715).

If a notification template was created by mistake or notification recipients no longer need to be notified, you can delete the template. You can delete a notification template that is not specified in any [response rule](#7481272715).

This section provides instructions on how to manage notification services and templates and a description of notification template settings.

### Managing email notification services and templates

You can set up email notifications about threats detected and incidents registered during monitoring of runtime events. For example, notifications about detected threats will be sent to employees of one department and notifications about registered incidents will be sent to employees of another department. For this, you need to deploy an SMTP server. In Runtime Radar, you need to set up a connection to it by adding a notification service and one or more notification templates for the service. You can use several SMTP servers if you add a separate notification service and a set of templates for each server.

**Adding an email notification service**

To add a notification service:

1. On the main menu, select **Notification services**.

1. In the top right corner of the page, select the cluster to add the notification service to.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Click **New service**.

1. Select the **Email** type.

1. Enter a name for the notification service.

1. Select an authentication type on the SMTP server.

1. Enter the LDAP server address and port in the following format: `<IP address or domain name>:<port>`.

1. Enter the account credentials to be used for authentication on the server.

1. If necessary, specify a root certificate of the certification authority.

1. Enter the email address of the sender on whose behalf notifications will be sent.

1. If your SMTP server supports secure connections with the STARTTLS extension, select the **Use STARTTLS** checkbox.

1. If you need to verify that the server connection is secure using a self-signed certificate, select the **Verify connection security** checkbox.

1. Click **Connect**.

   ***Note.** You can click **Connect without checking** to add the service without connecting to the SMTP server. For example, you might not have access to it yet.*

After adding an email notification service, you need to specify the addresses to notify via this service. To do this, you must create one or more notification templates. One template can include, for example, specialists from the same division of your organization.

**Adding an email notification service template**

To add a notification template:

1. On the main menu, select **Notification services**.

1. In the top right corner of the page, select the cluster for whose notification service to add the notification template.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Click ![pic](pics/9810272139.svg) or **New template** in the group box with the notification service for which you want to add a template.

1. Enter a name for the template.

1. Enter one or more recipients' addresses.

   ***Note.** The **Cluster** box displays the cluster for which the notification template will be created. The value matches the value selected in the **Cluster** list in the top right corner of the **Notification services** page.*

1. If necessary, change the central cluster address.

   ***Note.** The central cluster address is used to generate a link to an event in a notification. The **Central cluster address** box is filled in automatically. If the central cluster is accessed in a non-standard way, for example, the cluster is installed in an isolated network, you can specify the required address to access it.*

1. In the **Event type** list, select the event for notifications using this template.

1. Enter the message subject to be displayed to recipients.

1. If required, clear **Send default notification**.

   The **Notification body** box appears where you can add your own notification template using [syntax of the Go template processor](https://pkg.go.dev/text/template). Runtime Radar has an individual set of variables for each event type.

1. Click **Add**.

**Adding or excluding a recipient from notifications**

To add or exclude a recipient from notifications:

1. On the main menu, select **Notification services**.

1. In the top right corner of the page, select the cluster for which to add a recipient to a mailing list or exclude a recipient from the list.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Add or remove one or more email addresses.

1. Click **Save**.

### Managing webhook notification services and templates

You can configure webhook notifications about object scan and action check results.

Webhook is a method of integration in which an initiating system sends a notification to the recipient in the form of an HTTP request that contains all the necessary information.

This type of notifications requires a service that supports receiving of messages using webhook. This can be CI/CD systems (GitLab or Jenkins) or SOAR systems (Security Orchestration, Automation, and Response). You can also deploy a general-purpose webhook server that will perform the necessary operations.

In Runtime Radar, you need to set up a connection to the webhook server by adding a notification service and one or more notification templates for the service. You can implement integration with several webhook servers if you add a separate notification service and a set of templates for each server.

**Adding a webhook notification service**

To add a notification service:

1. On the main menu, select **Notification services**.

1. In the top right corner of the page, select the cluster to add the notification service to.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Click **New service**.

1. Select the **Webhook** type.

1. Enter a name for the notification service.

1. Enter a URL of the server to be used for notifications, and the port (if any) in the following format: `https://<IP address or domain name>:<port>`.

1. If the webhook server uses the basic authentication, enter the credentials to be used for authentication on the server.

1. If necessary, specify a root certificate of the certification authority.

1. If you need to verify that the server connection is secure using a self-signed certificate, select the **Verify connection security** checkbox.

1. Click **Connect**.

   ***Note.** You can click **Connect without checking** if you want to add the service without connecting to the webhook server. For example, you might not have access to it yet.*

After adding a webhook notification service, you need to specify the paths (endpoints) to notify. To do this, you must create one or more notification templates. For example, one template can contain a path for sending image scan notifications to a Mattermost group; the other template, a path for sending configuration scan notifications to another group.

**Adding a webhook notification template**

To add a notification template:

1. On the main menu, select **Notification services**.

1. In the top right corner of the page, select the cluster for whose notification service to add the notification template.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Click ![pic](pics/9810272139.svg) or **New template** in the group box with the notification service for which you want to add a template.

1. Enter a name for the template.

   ***Note.** The **Cluster** box displays the cluster for which the notification template will be created. The value matches the value selected in the **Cluster** list in the top right corner of the **Notification services** page.*

1. If necessary, change the central cluster address.

   ***Note.** The central cluster address is used to generate a link to an event in a notification. The **Central cluster address** box is filled in automatically. If the central cluster is accessed in a non-standard way, for example, the cluster is installed in an isolated network, you can specify the required address to access it.*

1. In the **Event type** list, select the event for notifications using this template.

1. Enter the path (endpoint) to which a webhook request will be sent.

1. If required, clear **Send default notification**.

   The **Notification body** box appears where you can add your own notification template using [syntax of the Go template processor](https://pkg.go.dev/text/template). Runtime Radar has an individual set of variables for each event type.

1. Specify keys and values for HTTP headers. You can add any number of headers.

1. Click **Add**.

### Managing syslog notification services and templates

By configuring syslog notifications about object scan and action check results, you enable integration with third-party systems for centralized storing and processing of security events.

You can use notifications received, for example, to normalize security events.

In Runtime Radar, you need to set up a connection to a system that supports the syslog protocol, or to a syslog server by adding a notification service and one or more notification templates for the service. You can implement integration with several systems that support the syslog protocol if you add a separate notification service and a set of templates for each system.

**Adding a syslog notification service**

To add a notification service:

1. On the main menu, select **Notification services**.

1. In the top right corner of the page, select the cluster to add the notification service to.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Click **New service**.

1. Select the **Syslog** type.

1. Enter a name for the notification service.

1. Enter the syslog server address and port in the following format: `<protocol>://<IP address or domain name>:<port>`. As a protocol, you can type `tcp` or `udp`.

1. Click **Connect**.

   ***Note.** You can click **Connect without checking** if you want to add the service without connecting to the syslog server. For example, you might not have access to it yet.*

When adding a syslog notification service, you must create one or more notification templates that will be used to generate notification content. The notification body format may depend on the system to which the notification will be sent. The standard configuration format complies with RFC 3164.

**Adding a syslog notification template**

To add a notification template:

1. On the main menu, select **Notification services**.

1. In the top right corner of the page, select the cluster for whose notification service to add the notification template.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Click ![pic](pics/9810272139.svg) or **New template** in the group box with the notification service for which you want to add a template.

1. Enter a name for the template.

   ***Note.** The **Cluster** box displays the cluster for which the notification template will be created. The value matches the value selected in the **Cluster** list in the top right corner of the **Notification services** page.*

1. If necessary, change the central cluster address.

   ***Note.** The central cluster address is used to generate a link to an event in a notification. The **Central cluster address** box is filled in automatically. If the central cluster is accessed in a non-standard way, for example, the cluster is installed in an isolated network, you can specify the required address to access it.*

1. In the **Event type** list, select the event for notifications using this template.

1. If required, clear **Send default notification**.

   The **Notification body** box appears where you can add your own notification template using [syntax of the Go template processor](https://pkg.go.dev/text/template). Runtime Radar has an individual set of variables for each event type.

1. Click **Add**.

### Notification template parameters

You can create a template for each notification. For this, Runtime Radar supports Go templates. You can create notification templates in various formats. For example, in JSON format for a webhook and syslog notification service, in HTML format for an email notification service, or in another format. The following table lists the available variables.

<table><caption>Variables available in the template</caption><colgroup><col style="width: 35.1%;"/><col style="width: 24.1%;"/><col style="width: 40.6%;"/></colgroup><thead><tr><th align="left">

Variable name
</th><th align="left">

Data type
</th><th align="left">

Description
</th></tr></thead><tbody><tr><td align="left">

`.notificationName`
</td><td align="left">

String
</td><td align="left">

Name of a notification template for which a notification is created
</td></tr><tr><td align="left">

`.centralCSURL`
</td><td align="left">

String
</td><td align="left">

Central cluster address 
</td></tr><tr><td align="left">

`.csClusterID`
</td><td align="left">

String
</td><td align="left">

ID of the cluster in which the event occurred
</td></tr><tr><td align="left">

`.csClusterName`
</td><td align="left">

String
</td><td align="left">

Name of the cluster in which the event occurred
</td></tr><tr><td align="left">

`.ownCSURL`
</td><td align="left">

String
</td><td align="left">

Address of the cluster in which the event occurred. For the central cluster, the value will be the same as for `.centralCSURL`
</td></tr><tr><td align="left">

`.event.GetSeverity`
</td><td align="left">

String
</td><td align="left">

Action severity
</td></tr><tr><td align="left">

`.event.GetRegisteredAt`
</td><td align="left">

[Google protobuf timestamp](https://protobuf.dev/reference/protobuf/google.protobuf/#timestamp)
</td><td align="left">

Event time
</td></tr><tr><td align="left">

`.event.GetBlock`
</td><td align="left">

Boolean
</td><td align="left">

Whether the action was blocked
</td></tr><tr><td align="left">

`.event.GetRuleName`
</td><td align="left">

String
</td><td align="left">

Name of a rule on whose triggering the notification is sent
</td></tr><tr><td align="left">

`.event.GetEventId`
</td><td align="left">

String
</td><td align="left">

Event UUID
</td></tr><tr><td rowspan="3" align="left">

`.event.GetEvent`
</td><td align="left">

String
</td><td align="left">

Object of the `Event` type that includes parameters of the String, *UInt32, and []String types. Parameters of the String type are as follows:

* `GetEventType`. The event type. Possible values: `PROCESS_EXEC`, `PROCESS_EXIT`, `PROCESS_KPROBE`, `PROCESS_TRACEPOINT`, `PROCESS_LOADER`, `PROCESS_UPROBE`.
* `GetPodNamespace`. The namespace where the event was registered.
* `GetPodName`. The name of the pod where the event was registered.
* `GetContainerName`. The name of the container where the event was registered.
* `GetContainerId`. The identifier of the container where the event was registered.
* `GetContainerImage`. The image from which the container was started.
* `GetFunctionName`. The name of the function intercepted using the kprobes tool. It will be populated if the event type is `PROCESS_KPROBE`.
* `GetProcessBinary`. The name of the executable file.
* `GetProcessArguments`. The arguments with which the executable file was run.
* `GetNodeName`. The name of the cluster node where the event was registered.
* `GetParentBinary`. The parent process executable file.
* `GetParentArguments`. The parent process arguments.
* `GetThreats`. The list of threats detected by detectors</td></tr><tr><td align="left">

*UInt32
</td><td align="left">

Parameters of the *UInt32 type:

* `ProcessPid`. The PID of the process.
* `ProcessUid`. The UID of the process.
* `ParentPid`. The PID of the parent process.
* `ParentUid`. The UID of the parent process.
* `ProcessSetuid`. The effective UID of the process. The parameter value will be something other than `NULL` if the executable file has the `setuid` access flag.
* `ProcessSetgid`. The effective GID of the process. The parameter value will be something other than `NULL` if the executable file has the `setgid` access flag</td></tr><tr><td align="left">

[]String
</td><td align="left">

[]String type parameters are as follows:

* `GetProcessCapEffective`. The list of effective capabilities of the process.
* `GetProcessCapPermitted`. The list of permitted capabilities of the process</td></tr><tr><td align="left">

`.event.GetEvent` →` GetThreats`
</td><td align="left">

String
</td><td align="left">

List of objects of the `Threat` type with the following structure:

* `GetDetectorId`. The identifier of the detector that detected the threat.
* `GetDetectorName`. The name of the detector that detected the threat.
* `GetDetectorVersion`. The version of the detector that detected the threat.
* `GetDetectorDescription`. The description of the detector that detected the threat.
* `GetSeverity`. The threat severity</td></tr></tbody></table>


Notification example:

```
<!doctype html>
<html lang="ru">
  <head>
    <meta charset="UTF-8" />
  </head>
  <body>
    <h4>Triggered rule: <b>{{.event.GetRuleName}}.</b></h4>
    <p>
      Time:
      <code>{{.event.GetRegisteredAt.AsTime.Format "02 Jan 06 15:04 MST"}}</code>
    </p>
    {{$event := .event.GetEvent}}
    <p>
      Suspicious activity was detected on the process startup
      <code>{{$event.GetEvent.GetProcessBinary}} {{$event.GetEvent.GetProcessArguments}}</code> in the pod
      <code>{{$event.GetPodNamespace}}.{{$event.GetPodName}}</code> on the host
      <code>{{$event.GetNodeName}}</code>.
    </p>
    <h4>Detected threats:</h4>
    <ul>
      {{range $t := $event.GetThreats}}
        <li>
          <b>{{$t.GetDetectorName}} ({{$t.GetDetectorId}})</b>:
          {{$t.GetDetectorDescription}}
        </li>
      {{end}}
    </ul>
    <p>
      To investigate, go to:
      <a href="https://<your server address>:3200/runtime/events">`Runtime Radar` console</a>
    </p>
    <p>
      <a href="{{.centralCSURL}}/events/{{.event.GetEventId}}?clusterUrl={{.ownCSURL}}">Go to the event</a>
    </p>
  </body>
</html>
```

> **Warning.** If the following parameters are used in a syslog or webhook notification template: `.event.GetEvent.ProcessPid`, `.event.GetEvent.ProcessUid`, `.event.GetEvent.ParentPid`, `.event.GetEvent.ParentUid`, `.event.GetEvent.ProcessSetuid`, or `.event.GetEvent.ProcessSetgid`, you need to implement checks for whether these parameters have values. Otherwise, the notifications based on such a template cannot be generated.

Example of checking whether a value is set:

```
{{if $args := $event.GetFunctionArgs}}
  "function_args": {{$args}},
{{end}}
```

## <a name="7481272715"></a>Managing response rules

Response rules determine how to respond when threats or vulnerabilities are detected in the specified objects.

You can manage response rules in the Runtime Radar web interface on the **Response rules** page, which opens as you select the **Rules** section on the main menu.

***Note.** Which features are available on the page depends on the user role.*

The page displays the following elements:
* **Cluster** list to select a cluster to manage. If child clusters are not yet connected to Runtime Radar or they are not yet in the **Installed** status, you can select only the **Central** option.
* Box to search for rules by name
* **Type** list to filter by rule types
* **Block** list to filter rules by the vulnerability severity that must be reached to block an image
* **Notify** list to filter rules by the vulnerability severity that must be reached to send a notification
* List of rules where you can view the details of each rule by clicking it
* **Create** button to create a response rule
* Buttons to edit or delete each rule

**Creating a rule for monitoring of and response to runtime events**

To set up automatic responses to the detection of threats during runtime event monitoring, you must create a response rule.

When you specify a pod name with a digest in the specification, the pod name can be displayed in different ways. This depends on the container runtime used. For example, a pod manifest contains the following image:

```
registry.k8s.io/ingress-nginx/controller@sha256:d56f135b6462cfc476447cfe564b83a45e8bb7da2774963b00d12161112270b7
```

If Containerd is used, the `image` parameter contains only the image ID hash (in `sha256:<hash>`), without a registry, path, or digest. Example: 

```
sha256:2d37f5a3dd01b3f22912802cdcbf8739693d2774f7e9d7c6f704ae3bd34fa0b0
```

If CRI-O is used, the `image` parameter contains a full link to the image, including the registry, path, and digest. Example:

```
registry.k8s.io/ingress-nginx/controller@sha256:62b61c42ec8dd877b85c0aa24c4744ce44d274bc16cc5d2364edfe67964ba
```

Runtime Radar receives data about containers from the status generated by the container runtime. This is due to the specifics of the container runtime implementation. Incorrect configuration of rules (without taking into account the runtime in use) may result in false positives or missed events. We recommend that you explicitly check the `image` value format in your runtime and take it into account when creating rules.

To create a rule for monitoring of and response to runtime events:

1. On the main menu, select **Rules**.

1. In the top right corner of the page, select the cluster for which to create the rule.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Click **Create**.

1. Enter a unique name for the response rule.

1. Enter the parameter values based on the container runtime in use.

   ***Note.** In name templates, you can use an asterisk (`*`) for any number of any character and a question mark (`?`) for any single character. For example, the `default-*` template matches the namespaces whose names start with `default`. Templates are case-sensitive. Regular expressions are not supported.*

1. Select the threat severity that must be reached to block a runtime event.

   ***Note.** An event is blocked by deleting the pod where the event occured.*

1. Select the threat severity that must be reached to send a notification.

1. Select a notification template.

   If **Do not notify** is selected under **Notify**, the templates are not available.

   ***Note.** You can select a template if at least one notification template is created in Runtime Radar.*

1. If required, under **Exclusions**, select threats for the rule to skip.

1. If required, under **Exclusions**, enter paths or templates of paths to executable files for the rule to skip.

1. Click **Create**.

## Monitoring of and responding to runtime events

You can set up monitoring of and responding to runtime events. Monitoring of runtime events allows tracking of events at the level of individual pods or containers in Kubernetes clusters, including the start of processes, system calls, and requests to specific kernel functions. During monitoring of and responding to events, they are checked through a chain of detectors that detect threats in an event and assign a severity to them. Response rules created in Runtime Radar allow you to configure responses to be performed when a threat is detected.

A detector is a program written in a Turing-complete programming language (for example, Go, Python, C, C++, Rust) and compiled in the WebAssembly (Wasm) format; it can be dynamically loaded to the system and unloaded without restarting, and allows you to implement secure algorithms, which are close in speed to native code, for checks and analysis of incoming events.

You can configure tracking of runtime events from specific sources to get check results of only required events and reduce the flow of events processed by Runtime Radar.

A source is a description of operation logic for eBPF programs in a special language (TracingPolicy). The programs are loaded to and executed in an OS kernel on a host and can track runtime events using the `kprobe`, `uprobe`, `tracepoint`, and `BPF-LSM` mechanisms of the Linux kernel. Sources use the mechanisms described above to track the activity of processes in containers (monitoring of system calls and requests to specific kernel functions) and detect threats. Events that start and stop processes in cluster containers are always tracked. Using the expert mode, you can edit existing sources or add new sources.

> **Warning.**[Configure sources and event filters](#7786230795) taking into account the actual load on protected clusters to avoid excessive load on Runtime Radar. If the load increases, consider scaling the system.

This section describes how Runtime Radar monitors and responses to runtime events, and provides instructions on how to manage such events in the system.

The process of monitoring of and responding to runtime events comprises the following stages:
1. Checking runtime events through a chain of detectors and setting a severity for detected threats.
1. Saving the events to the database depending on the event logging settings. Saved events are displayed on the **Events** tab of the **Runtime** page.
1. Checking the events according to the response rules created in Runtime Radar to define further responses.

If a response rule is triggered by the event check, an incident is registered in the system.

You can manage monitoring of and response to runtime events in the Runtime Radar web interface on the **Runtime** page, which opens as you select the **Runtime** section on the main menu.

***Note.** Which features are available on the page depends on the user role.*

In the top right corner of the page, there is the **Cluster** list for selecting a cluster to manage. If child clusters are not yet connected to Runtime Radar or they are not yet in the **Installed** status, you can select only the **Central** option.

The page contains the following tabs:
* **Parameters** to [manage event sources, the allow and deny filters, as well as configure event logging](#7786230795).
* **Events** to [view information about events](#5265696395) that were logged by Runtime Radar according to the values on the **Parameters** tab.
* **Detectors** to view information about detectors, [add, and delete them](#7163413643).

### <a name="9979340811"></a>Managing runtime event monitoring sources

The Runtime Radar distribution kit includes a set of sources covered by the Runtime Radar license. The sources from the distribution kit and the types of events they track are described in the table below.

<table><caption>Sources and types of events that they track</caption><colgroup><col style="width: 50.0%;"/><col style="width: 50.0%;"/></colgroup><thead><tr><th align="left">

Source name
</th><th align="left">

Types of tracked events
</th></tr></thead><tbody><tr><td align="left">

Outgoing TCP connections
</td><td align="left">

This source tracks the `tcp_connect`, `tcp_close`, and `tcp_sendmsg` functions, allowing detection of outgoing TCP connections (including the connection establishment, termination, and sending of TCP packets). Enabling this source can significantly increase the event flow and load on the system. In this case, you should narrow the flow using more accurate filters (for example, specify only certain pods)
</td></tr><tr><td align="left">

Privilege escalation
</td><td align="left">

This source tracks the `commit_creds` function allowing detection of privilege escalation, including superuser (root) privileges
</td></tr><tr><td align="left">

Using tools for debugging and reverse engineering (`ptrace`)
</td><td align="left">

This source tracks the `ptrace` system call, which may indicate attacker activity in the target system
</td></tr><tr><td align="left">

Start and stop of processes
</td><td align="left">

This source tracks the start and stop of processes in containers on nodes with installed runtime monitoring components. Information about the process context that an event occurs in is added to other types of events (for example, `kprobe`) from other sources, so this source cannot be disabled. Its impact on the overall system load is usually small
</td></tr><tr><td align="left">

Loading and unloading of kernel modules
</td><td align="left">

This source tracks the `do_init_module`, `free_module`, `security_kernel_module_request`, and `security_kernel_read_file` calls, allowing detection of explicit or implicit (automatic) loading and unloading of modules, as well as attempts to manipulate modules and other malicious activity
</td></tr><tr><td align="left">

Opening of a socket for incoming connections
</td><td align="left">

This source tracks the `inet_csk_listen_start` call, revealing possible activity of unwanted networking tools. The source also detects legitimate activity in a container; therefore, in addition to opening of a socket, detectors also track other event parameters
</td></tr><tr><td align="left">

Device mounting
</td><td align="left">

This source tracks the `do_mount` call, allowing detection of potentially unwanted events related to device mounting
</td></tr><tr><td align="left">

Copying of file descriptors
</td><td align="left">

This source tracks calls to functions that copy file descriptors. The source tracks the copying of the standard input file descriptor (stdin), which may indicate an attempt to build a pipe required for various hacking tools
</td></tr><tr><td align="left">

Monitoring the `io_uring` interface
</td><td align="left">

The source tracks the `io_uring_setup()` and `io_uring_enter()` system calls, which indicate the use of the `io_uring` interface in the system
</td></tr><tr><td align="left">

Use of usermode helper API
</td><td align="left">

The source tracks calls of the Linux kernel functions `call_usermodehelper_setup()` and `call_usermodehelper_exec()`, which indicate process configuration and startup via the usermode helper API. This helps detect several different attacks that use the specified interface to run processes at the host OS level
</td></tr><tr><td align="left">

Rootkit upload monitoring
</td><td align="left">

The source tracks calls of the Linux kernel function `kallsyms_lookup_name()` to retrieve the address of a system call table, which may indicate rootkit activity in the target system. The source tracks if eBPF programs were loaded using the `bpf_check()` function, which may help search for rootkits that use eBPF
</td></tr></tbody></table>


**Managing sources**

Using the expert mode, you can edit existing sources or add new sources. The expert mode is available only to users with the **Administrator** role.

You can manage sources on the **Runtime** page on the **Parameters** tab under **Sources**. Under **Sources**, the **Expert mode** switch and a list of sources are displayed. For each source, a checkbox is displayed to enable or disable event tracking from this source. You can open a window with information about a source by clicking its line. The window displays the name, description, and source code. For source code, the following buttons are displayed:
*  to enable word wrapping.
*  to copy source code.
*  to save code to a text file.

In the expert mode, the **Add** button is available for adding a source. If a source is edited or added, the **Reset to default** button appears. It restores the original state of event sources and filters. If [Grafana integration and monitoring](#9502135051) of Runtime Radar components are set up, then, after you edit or add sources, you can view the metrics.

To add a source for runtime event monitoring:

1. On the main menu, select **Runtime**.

1. In the top right corner of the page, select the cluster to add the source to.

1. Enable the expert mode.

1. Click **Add**.

1. Enter the name, description, and source code.

   ***Note.** Source code will be checked automatically. If an error is detected, a message with the number of the first line containing errors will appear.*

1. Click **Add**.

1. Click **Apply**.

The added source will be the last in the list. To start monitoring events from a new source, you must enable it.

***Note.** You cannot edit or disable the **Start and stop of processes** source. This source tracks the start and stop of processes in containers on nodes with installed runtime monitoring components and adds the process context to events of other types (for example, kprobe) from other sources.*

You can delete the source by clicking , which appears when you hover over the source line.

### <a name="7786230795"></a>Managing tracking settings and event filters

**Setting up tracking of events**

You can set up tracking of runtime events. This allows you to get check results of only required events, and reduce the load on the system components.

To set up tracking of events:

1. On the main menu, select **Runtime**.

1. In the top right corner of the page, select the cluster for which to set up tracking.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Select [the sources](#9979340811) of runtime events you want to track.

   ***Note.** You can view the detailed source description by clicking the line with its name.*

1. Add allow and deny event filters.

1. Select the event logging option:

   * To log all events and display them on the **Events** tab, select **All**.

   ***Note.** Logging of all events allows for full investigation, but to avoid excessive load on Runtime Radar, you should configure event filters and sources, taking into account the actual load on protected clusters. If the load increases, consider scaling the system.*

   * To log only events where threats were detected and display them on the **Events** tab, select **With threats only**.

   ***Note.** The **With threats only** option limits the incident investigation possibilities. You can select it when the database fails to handle the load with the parameters specified under **Event tracking** and **Lists of event filters**.*

   * To keep responding to events without logging the events and displaying their details on the **Events** tab, select **Do not log**.

   ***Note.** The **Do not log** option disables the incident investigation.*

1. Click **Apply**.

**Managing event filters**

You can add allow and deny event filters to configure the flow of events coming to Runtime Radar for monitoring and response.

Adding an allow filter means that events from specified namespaces and pods with labels specified will be processed in Runtime Radar by a chain of detectors. Events that do not match the filter criteria will not be processed.

If multiple allow filters are added, an event to be processed must match conditions of at least one of the filters. If several deny filters are added, the event is excluded from the check if it matches conditions of at least one of the filters.

If allow filters and deny filters are added, the event is first checked against the allow filters. If it matches at least one of them, the event is checked against the deny filters, and then the event is processed by a chain of detectors. If the event does not match the allow filters or matches both the allow filters and deny filters, the event is not processed.

To add an event filter:

1. On the main menu, select **Runtime**.

1. In the top right corner of the page, select the cluster to add the event filter to.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Click **Add filter**.

1. Select which filter you want to add: **Allow filter** or **Deny filter**.

1. Specify namespaces for event filtering.

1. Specify regular expressions to identify pods. For the regular expression format, see the [Go documentation website](https://pkg.go.dev/regexp/).

   ***Note.** For example, the `news-service` regular expression as a plain string without additional characters and constructs will match the names of pods that contain the string anywhere in the name.*

1. Enter labels of the pods in the key=value format. The following operators are allowed between key and value: `=`, `==`, and `!=`. For the format of label keys and values, see [the Kubernetes documentation](https://kubernetes.io/docs/).

   ***Note.** For example, if the `app.kubernetes.io/name=tetragon` value is set, all pods will be filtered whose label key is `app.kubernetes.io/name` and value is `tetragon`.*

1. Click **Add**.

1. Click **Apply**.

You can edit the filter by clicking ![pic](pics/9777620363.svg) and delete the filter by clicking ![pic](pics/9697536907.svg). At least one event filter is required. Without filters, too many events will be tracked. After the filter is deleted, Runtime Radar will track events that match the filter, without using conditions.

### <a name="7163413643"></a>Managing detectors

The Runtime Radar distribution kit includes a set of detectors covered by the Runtime Radar license. Customers, third-party vendors, and community members also can develop detectors and upload them to the system.

You can manage detectors on the **Runtime** page on the **Detectors** tab. Information about all detectors (the identifier, name, version, and description) is presented as a table.

**Adding a detector**

To add a detector:

1. On the main menu, select **Runtime**.

1. Go to the **Detectors** tab.

1. If necessary, in the top right corner of the page, select the cluster to add the detector to.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Click **Add**.

1. Select or drag files to upload.

   ***Note.** The detector must be written in a Turing-complete programming language (for example, Go, Python, C, C++, Rust) and compiled in the WebAssembly (WASM) format.*

1. Click **Add**.

You can delete a detector by clicking ![pic](pics/9697536907.svg).

### <a name="5265696395"></a>Viewing information about runtime events

You can view the results of runtime event monitoring according to the specified [event monitoring parameters](#7786230795) by clicking the **Events** tab on the **Runtime** page. To open this page, click **Runtime** on the main menu.

In the top right corner of the page, the page displays the **Cluster** list for selecting a cluster and the ![pic](pics/9854071947.svg) button for refreshing the event table. If child clusters are not yet connected to Runtime Radar or they are not yet in the **Installed** status, you can select only the **Central** option.

The page also displays buttons for filtering events and the runtime event table.

The following filtering options are available:
* By parameters and their values. You can also use this option to filter events with threats by detectors and events with incidents by triggered rules.
* By context. This option is linked to an event from the list and allows you to filter events by their relationship to a parent event. Possible values:
   * Parent context. Display events related to the parent process of the selected event.
   * Same parent. Display events related to the same parent process as the selected event.
   * Child process. Display all child events related to the same process as the selected event.
   * Same process. Display events related to the same process as the selected event.
* Preset filters. You can select one of the preset filters.
* Last filters. The filter settings are saved to the system so you can use the previous search.

The event table displays the following:
* Event parameters (namespace/pod, function, executable file, argument, and event date).
* Number of detected threats and their severity. You can view the list of detected threats by clicking the threat severity icon.
* ![pic](pics/9646986891.svg) to view raw event data.
* ![pic](pics/9854331275.svg) to select a filter by context.

In addition, an event row may display the ![pic](pics/9854339979.svg) icon if threats are detected and there are triggered rules or ![pic](pics/9855419147.svg) if threats are detected but no rules are triggered. If no threats are detected, the icons are not displayed. Clicking ![pic](pics/9855419147.svg) opens a sidebar for creating a response rule. The **Namespaces** box and the minimum threat severity for sending notifications are automatically filled in with the data of the event for which the rule is created. The threat severity is the maximum severity for the detected threats.

Rows with events whose check failed with errors are highlighted in the table in red. You can view the list of detectors that failed to complete the check and related errors by clicking the threat severity icon in the row.

Clicking one of the events brings you to the viewing page of the separate event.

The page of a separate runtime event displays the following elements:
* Event ID and date.
* Group box with parameters of the event's parent process.
* Group box with event process parameters.
* ![pic](pics/9646986891.svg) to view raw event data.
* ![pic](pics/9854331275.svg) to select a filter by context.
* List of events with the same parent process as the current event.
* Side panel to the right with incident parameters (if it is registered) and the following lists:
   * Triggered rules that can be edited and deleted.
   * Detected threats.
   * Detectors that failed to complete the check because of an error.

## <a name="9502135051"></a>Integrating Runtime Radar with Grafana

[Grafana](https://grafana.com/docs) is a data visualization system. Grafana displays information in a format that is easy to analyze and can be used to monitor the operation of Runtime Radar.

For integration, you must set up a connection to Grafana when [installing](#9815029643) or updating Runtime Radar in the central cluster. You can also set up a connection to Grafana when installing Runtime Radar in [child clusters](#9839873547).

After the installation, the Grafana web interface will be available at `<address of theRuntime Radarweb interface in the central cluster>/grafana`. To log in, enter the credentials specified during the Grafana installation.

**Dashboards in Grafana**

A dashboard in Grafana is a page with charts, diagrams, and other statistics about operation of a particular IT system.

Grafana provides a set of dashboards for monitoring Runtime Radar. Each dashboard displays specific information about the state of system components and services. On each dashboard, you can select a cluster for which to display the information.

The following dashboards are included in the distribution kit:
* Admission. For monitoring detectors used for checks via the admission controller. Shows the number of operation errors and duration of detector chain execution.
* Cluster status. Provides information about the health of ClickHouse, RabbitMQ, and PostgreSQL, the current cluster state (central cluster, unregistered child cluster, and registered child cluster), and the Runtime Radar license and vulnerability database update status. To get information about the health of external services, you must configure the internal Prometheus system for this purpose or install the external Prometheus system.
* Go app. Provides the main metrics for Golang services. Shows the key performance and stability indicators for these components.
* gRPC. Gives a summary of gRPC communications between Runtime Radar services (request types and their statuses and duration). The majority of internal communications between services is done via gRPC. To make analysis easier, all related metrics are gathered on one dashboard.
* Public API. For monitoring the public API. Shows the external interface health and performance to help control its availability and quality of provided services.
* Reverse proxy. Provides information about the availability and operation of the reverse proxy component. All internal and external requests to Runtime Radar are routed through the reverse proxy component. The dashboard allows you to detect anomalies in the service and promptly react to possible issues.
* Runtime. For managing metrics of services dedicated to the in-cluster deployment. If you use external components, such as RabbitMQ or ClickHouse, for correct filtering by the **Cluster** parameter and if there is Grafana configuration for several clusters, add the "cluster" label to the metrics of these components (similar to other system metrics). You can also view information about the runtime monitor service connection to RabbitMQ.

Each dashboard contains annotations and notifications that help promptly respond to changes in the component state.

If you use Grafana from the Runtime Radar distribution kit, dashboards will be available in the Grafana web interface. If you use an external Grafana, you need to import the preset dashboards for monitoring Runtime Radar. The dashboard files are available in the `dashboards` directory of the installation package archive with the Helm chart.

To import the Runtime Radar dashboard to Grafana:

1. Log in to the Grafana web interface.

1. On the main menu, click ![pic](pics/281055243.svg) and select **Dashboards**.

1. Click **New** → **Import**.

1. Click **Upload dashboard JSON file**.

1. Select the Runtime Radar dashboard file.

1. Click **Load**.
