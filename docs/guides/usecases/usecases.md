# 
## Use cases

This section describes the most popular Runtime Radar use cases.

### Integration of container runtime monitoring into SOCs

Integration of container runtime monitoring into SOCs allows you to detect threats in real time and send information about them to SIEM systems.

Integration of container runtine monitoring into SOCs has the following advantages:
* Threat detection in real time. Suspicious actions are logged instantly, reducing the risk of incidents being overlooked.
* Analysis automation and response time reduction. Incident information is automatically transferred to SIEM systems. SOC analysts receive structured messages with event information, which reduces the investigation time and allows them to quickly block attacks. You can create notification templates in various formats. For example, in JSON format for a webhook and syslog notification service, in HTML format for an email notification service, or in another format.
* Increased visibility in modern infrastructures. This provides a deep analysis of container behavior based on the monitoring of application calls to the operating system core. This is important for detecting hidden threats, such as backdoors or cryptominers.
* Data correlation in a single system. Integration with SIEM systems allows you to link events from the container environment to data from other sources (network devices, cloud services, and endpoints) in order to detect complex attacks.
* SOC load reduction. Automated exclusion of outdated threats and grouping of incidents according to business logic allow the team to focus on the most dangerous threats and improve efficiency.

To configure the integration, you must add a syslog notification service and its template and create response rules.

To add a notification service:

1. On the main menu, select **Notification services**.

1. Click **New service**.

1. Select the **Syslog** type.

1. Enter a name for the notification service.

1. Enter the syslog server address and port in the following format: `<protocol>://<IP address or domain name>:<port>`. As a protocol, you can type `tcp` or `udp`.

1. Click **Connect**.

1. Under the notification service, click ![pic](pics/9810272139.svg).

1. Enter a name for the template.

1. Click **Add**.

To create a response rule:

1. On the main menu, select **Rules**.

1. Click **Create**.

1. Enter a rule name.

1. Specify the rule settings.

1. Under **Notify**, select the vulnerability severity that must be reached to send a notification.

1. Select the template you specified when creating the notification service.

1. Click **New**.

### Automation of responding to information security incidents in container environments

You can automate responses to information security incidents in container environments using a webhook. Webhook is a method of integration in which an initiating system sends a notification to the recipient in the form of an HTTP request that contains all the necessary information.

Webhook integration has the following advantages:
* Instant incident response. A webhook allows you to start a workflow in real time when threats are detected.
* Integration with existing systems without infrastructure modification. It can be easily integrated with SOC platforms, SIEM solutions, or monitoring tools. This allows automated processing of security events without refactoring of the cluster architecture, as well as unification of processes across different systems.
* Lower risks caused by the human factor. Automation of standard actions excludes errors that may occur when data is entered manually. This is especially important in stressful scenarios when the SOC analyst must focus on the analysis of complex attacks.
* Scalability and flexibility. When expanding the cluster or adding new services, you can use the webhook mechanism to streamline automation management. For example, a single configuration point for all cluster nodes makes it easier to implement updated security policies without manually reconfiguring each component.
* SOC resource saving. Automation gives analysts more time to manage non-trivial threats, which increases the overall team efficiency and reduces load on personnel.

To configure the integration, you must add a notification service using a webhook and its template and create response rules.

To add a notification service:

1. On the main menu, select **Notification services**.

1. Click **New service**.

1. Select the **Webhook** type.

1. Enter a name for the notification service.

1. Enter a URL of the server to be used for notifications, and the port (if any) in the following format: `https://<IP address or domain name>:<port>`.

1. If the webhook server uses the basic authentication, enter the credentials to be used for authentication on the server.

1. Click **Connect**.

1. Under the notification service, click ![pic](pics/9810272139.svg).

1. Enter a name for the template.

1. Click **Add**.

To create a response rule:

1. On the main menu, select **Rules**.

1. Click **Create**.

1. Enter a rule name.

1. Specify the rule settings.

1. Under **Notify**, select the vulnerability severity that must be reached to send a notification.

1. Select the template you specified when creating the notification service.

1. Click **New**.

### Protection of multicluster environments

With the product, you can deploy all security agents in each infrastructure cluster and have centrilized control over container security. This reduces incident response time and makes it easier to investigate and analyze them.

You must first install Runtime Radar in the central cluster and then connect child clusters to the Runtime Radar central cluster. This allows you to do the following:
* Connect only the required security sensors in each of the child clusters to save resources.
* Use external queue managers, DBMS servers, caching systems, and storage spaces to reduce resource usage and infrastructure costs to collect and process large event flows.
* Simplify incident management in a complex infrastructure using a single management point and provide centralized monitoring and response to information security events.
* Ensure the maximum infrastructure coverage.
* Apply the common authentication mechanism for all clusters to ensure more reliable access control over sensitive information.

To protect the multicluster environment, you must do the following:
1. Connect child clusters in the Runtime Radar web interface.
1. Install Runtime Radar in child clusters.
1. Configure response rules in the Runtime Radar web interface.

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

**Configuring response rules**

To configure response rules, you must add at least one notification service and its template and create response rules.

To add a notification service:

1. On the main menu, select **Notification services**.

1. Click **New service**.

1. Select the service type.

1. Enter a name for the notification service.

1. Enter the settings for connecting to the notification service server.

1. Click **Connect**.

1. Under the notification service, click ![pic](pics/9810272139.svg).

1. Enter a name for the template.

1. Click **Add**.

To create a response rule:

1. On the main menu, select **Rules**.

1. Click **Create**.

1. Enter a rule name.

1. Specify the rule settings.

1. Under **Notify**, select the vulnerability severity that must be reached to send a notification.

1. Select the template you specified when creating the notification service.

1. Click **New**.
