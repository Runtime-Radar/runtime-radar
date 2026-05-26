# Quick start guide

## Quick installation using Helm

The Helm chart configuration file with the default settings will be used for installation. If you need to consider the specifics of the existing infrastructure, you can manually fill in the Helm chart configuration file prior to installation. All of the available settings are described in the [README.md](../../install/helm/README.md) file.

To install Runtime Radar using Helm,

1. Run the following command:

   ```bash
   helm install runtime-radar -n runtime-radar --create-namespace oci://ghcr.io/runtime-radar/runtime-radar:v0.1.0 \
     --set-string 'global.ownCsUrl=https://your-domain.com:32000' \
     --set-string 'global.keys.publicAccessTokenSalt=INIT-DO-NOT-USE' \
     --set-string 'global.keys.encryption=INIT-DO-NOT-USE' \
     --set-string 'global.administrator.username=admin' \
     --set-string 'global.administrator.password=Password' \
     --set-string 'reverse-proxy.service.type=NodePort' \
     --set-string 'reverse-proxy.service.nodePorts.http=32000'
   ```

   ***Note.** In the command example, the username is `admin` and the password is `Password`. You can specify other values and later use them to connect to the Runtime Radar web interface.*

   ***Note.** In the command example, access to the web interface is configured using the NodePort service on port 32000. You can use the Ingress controller instead or change the port number. To do this, you must specify the corresponding settings. You can also change other settings in the installation command.*

Now you can start [setting up the runtime event monitoring](#9807610635).

## <a name="9807610635"></a>Configuring monitoring of runtime events

Monitoring of runtime events allows tracking of events at the level of individual pods or containers in Kubernetes clusters, including the start of processes, system calls, and requests to specific kernel functions. During monitoring of and responding to events, they are checked through a chain of detectors that detect threats in an event and assign a severity to them. Response rules created in Runtime Radar allow you to configure responses to be performed when a threat is detected.

The process of monitoring of and responding to runtime events comprises the following stages:
1. Checking runtime events through a chain of detectors and setting a severity for detected threats.
1. Saving the events to the database depending on the event logging settings. Saved events are displayed on the **Events** tab of the **Runtime** page.
1. Checking the events according to the response rules created in Runtime Radar to define further responses.

**Configuring monitoring**

To configure event monitoring:

1. On the main menu, select **Runtime**.

1. Select the check boxes for all available runtime event sources.

1. Add allow and deny event filters.

1. Select the event logging option.

1. Click **Apply**.

**Configuring response rules**

To configure response rules, you must add at least one notification service and its template and create response rules.

To add a notification service:

1. On the main menu, select **Notification services**.

1. Click **New service**.

1. Select the service type.

1. Enter a name for the notification service.

1. Enter the settings for connecting to the notification service server.

1. Click **Connect**.

1. Under the notification service, click ![pic](pics/icon_add.svg).

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
