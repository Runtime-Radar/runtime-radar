# Developer help

## Detector Development Guide

The Runtime Radar container runtime protection system ensures continuous monitoring of events within the OS kernel and userspace with tracing tools (for example, kprobe, uprobe, and tracepoint). The eBPF technology and the open-source product [cilium/tetragon](https://github.com/cilium/tetragon/) are used to obtain events using the Linux kernel mechanisms. The Runtime Radar runtime monitoring component manages Tetragon instances, TracingPolicy uploads, and filters for restricting the event flow. These changes take effect without system restart.

Tetragon supports threat detection via TracingPolicy. However, Runtime Radar uses its own WebAssembly (Wasm) engine to search for threats in runtime events. This functionality is implemented in the event processor component. Wasm provides performance comparable to native code due to AOT compilation in the runtime software. Wasm modules operate in isolation to prevent detectors from accessing the main process memory and reduce the risk of components affecting each other. At the same time, Wasm runtime vulnerabilities are rare and hard to exploit, which increases the overall solution security. Due to Wasm being a cross-platform solution, compiled modules can run on any OS and architecture compatible with Wasm runtime, which provides deployment flexibility.

A detector is a Go program that searches for threats in runtime events. The general-purpose programming language is used, so you can implement the threat detection logic of any complexity.

**Detector code structure**

The detector code is logically divided into several sections:
1. The section with the detector metadata starts with the `const` keyword and contains the following parameters: `ID`, `Name`, `Description`, `Author`, `Contact` (author's contact information), and `Lisense`. After a detector is uploaded to Runtime Radar, this metadata will be displayed in the web interface.
   ```
   const (
     ID = "TEST_NETCAT_LISTEN"
     Name = "Use of netcat for creating incoming connections"
     Description = "The detector detects network activity related to the use of netcat to create incoming connections.
     Version = 1
     Author = "CS Team"
     Contact = "email: cs@example.com"
     License = "Apache License 2.0"
   )
   ```
1. The section with detector trigger criteria required for its optimization starts with the `var` keyword. Criteria allow you to trigger the detector only for specific event types and functions and reduce the event processing time. You can specify types of events the detector must process and function names (if relevant for the type). For event type details, visit [tetragon.io](https://tetragon.io/docs/reference/grpc-api/#eventtype). The detector can manage different event types and different functions simultaneously. When describing criteria, make sure that the specified functions are traced by the policies (TracingPolicy). If you want the detector to trigger on all events, you must specify all event types. The policies [provided](https://github.com/Runtime-Radar/runtime-radar/tree/main/runtime-monitor/pkg/model/tracingpolicy) together with the Runtime Radar current version support the `PROCESS_EXEC` and `PROCESS_KPROBE` event types. For the detector to trigger on all of the called functions, you must leave the list of functions empty or enter `*`. The detector example contains only the `inet_csk_listen_start` kernel function because the detector must not be triggered when other functions are called or when events of other types occur.
   ```
   var (
     // triggerCriteria sets trigger criteria as a map between event types and corresponding functions
     // to be used by the detector. If function names are not applicable to
     // a particular event type, such as "PROCESS_EXEC", leave the slice empty or use
     // the "*" wildcard.
     triggerCriteria = map[string][]string{
       "PROCESS_KPROBE": {"inet_csk_listen_start"},
       // Examples:
       //
       // "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
       // In order to process all possible functions, leave the right-hand part empty or use the "*" wildcard:
       // "PROCESS_EXEC": {},
       // same as:
       // "PROCESS_EXEC": {"*"},
     }
   )
   ```
1. The section with global variables starts with the `var` keyword.
   ```
   var ( nc = glob.MustCompile("*/nc") )
   ```
1. The section for selecting an action (switch-case) depending on the PROCESS_KPROBE event type starts with the `switch` keyword and contains multiple `case` sections to process various event types.
   ```
     switch ev := event.(type) {
     case *tetragon.GetEventsResponse_ProcessExec:
       // Nothing here
     case *tetragon.GetEventsResponse_ProcessExit:
       // Nothing here
     case *tetragon.GetEventsResponse_ProcessKprobe:
       kprobe := ev.ProcessKprobe
       binary := kprobe.GetProcess().GetBinary()
       function := kprobe.GetFunctionName()
   ```
1. The section for extracting from an event a path to the `binary` file and the `function` function name starts with the `if` keyword. In this section, data is extracted from the event. Tetragon events are described using generated protobuf stubs. This allows you to manage them in the module code in the same way as if you were working with these events outside WebAssembly in a service that interacts with Tetragon directly via gRPC. As with the regular `protobuf`, we recommend that you use getters wherever possible to avoid a panic when accidentally calling a `nil` value. The target function is simultaneously being checked if it matches the `inet_csk_listen_start` value and if it is correct. If the target function does not match `inet_csk_listen_start`, the detector operation is terminated with a zero result, that is, without a detected threat.
   ```
       if !(nc.Match(binary) && function == "inet_csk_listen_start") {
         return resp, nil
       }
   ```
1. The next section starts with the `args` variable. In this section, a slice with argument object indices (`args`) is extracted and checked that its length is above the required minimum value. If the condition is not met, the detector returns an empty result and a detailed error message generated using `fmt.Errorf`. We recommend that you add similar checks to all parts of the code where index references are intended to avoid a panic when getting slice elements by index. A panic inside a specific detector does not affect other detectors or the event processor component. However, it may lead to the detector failure. The required object is first in the list. You can get it by index `0`. Then, using the `GetSockArg().GetSport()` method chain, you can get the target port number. If the target port is 8888, the threat is assigned critical severity. Then, a result that confirms threat detection is returned.
   ```
       args := kprobe.GetArgs()
       if len(args) < 1 {
         return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
       }
       sport := args[0].GetSockArg().GetSport()
       if sport == 8888 {
         resp.Severity = api.DetectResp_CRITICAL
         return resp, nil
       }
       // Nothing here
     case *tetragon.GetEventsResponse_ProcessTracepoint:
       // Nothing here
     }
     return resp, nil
   }
   ```

**Setting up the environment**

To develop and compile detectors, you need to set up the environment and use Go version 1.25, TinyGo version 0.39.0, and an IDE that supports Go. We recommend that you use Visual Studio Code as your IDE. Configuration files for this development environment are in the `/event-processor/detector/wasm/.vscode` directory. Before you get started with detector code, you must install recommended extensions for Visual Studio Code and restart the IDE. To employ other IDEs, you must examine the contents of the `.vscode/settings.json` file and set up the environment in the same way.

To check the TinyGo version,

1. Run the following command:

   ```
   tinygo version
   ```

To get started with detector code,

1. Open the directory `event-processor/detector/wasm` in a separate instance of Visual Studio Code.

**Example of how to develop a detector**

For example, let us develop a detector that detects if `netcat` is used to listen to incoming connections on port 8888. This may be a sign of an attacker's lateral movement or attempts to build a reverse shell.

The protector must detect if the `inet_csk_listen_start` kernel function is started. You can find data about a TCP port for incoming connections in the `sock` structure a pointer to which is passed as an argument of a detected function. The corresponding source (TracingPolicy) is already in Runtime Radar. 

Runtime Radar has source code of detectors that process events of different types. All of them have a similar structure. Before you start developing a detector, you must find the most suitable detector for your scenario and use it as a basis. In this example, we will use the **CS_RT_SSH_TUNNEL_USE** detector code.

To develop a detector:

1. Copy the directory with the **CS_RT_SSH_TUNNEL_USE** detector using its `ID` as its name to a new directory, for example, **TEST_NETCAT_LISTEN**. 

1. In the `main.go` file, make changes in the required sections.

You can examine the [full source code](#9960038795) of the resulting detector.

**Compiling detectors**

To compile detectors, you need the [TinyGo](https://tinygo.org/getting-started/install) compiler, wasip1 target, and [Task](https://taskfile.dev/docs/installation) utility. The event processor module directory contains the `Taskfile.yml` file with compilation scripts. The current Runtime Radar version only supports detectors compiled by TinyGo version 0.39.0.

To compile all of the available detectors:

1. Go to the event-processor/detector/wasm directory in the product repository.

1. Run the compile command:

   ```
   task detectors
   ```

The compiled detector is now available in the `/event-processor/deploy/` directory.

You can compile a detector by running the TinyGo compiler manually in a new detector directory.

To compile the detector manually:

1. Go to the directory of the new detector:

1. Run the compile command:

   ```
   tinygo build -target=wasip1 -scheduler=none --no-debug -o <detector filename>.wasm
   ```

   Example:

   ```
   tinygo build -target=wasip1 -scheduler=none --no-debug -o TEST_NETCAT_LISTEN.wasm
   ```

**Adding a detector**

To add a detector:

1. On the main menu, select **Runtime**.

1. Go to the **Detectors** tab.

1. If necessary, in the top right corner of the page, select the cluster to add the detector to.

   ***Note.** If during Runtime Radar deployment, a self-signed certificate was used, to access the child cluster, you may need to follow the child cluster URL by adding the URL to the security exceptions, ignore the warning about an insecure connection, or add the certificate to the trusted certificates and then try to select a child cluster again.*

1. Click **Add**.

1. Select or drag files to upload.

1. Click **Add**.

**Checking a detector**

To check a detector, we recommend that you create a test pod and set up its monitoring.

To check a detector:

1. In the Runtime Radar interface on the main menu, select **Runtime**.

1. Make sure that the **Opening of a socket for incoming connections** source is enabled.

1. In the test pod, run the following command:

   ```
   nc -vv -l -p 8888
   ```

1. Run several simple commands to save the messages accumulated in the buffer in the DB. Example:

   ```
   ls -al
   ```

1. In the Runtime Radar interface on the main menu, select **Runtime**.

1. Go to **Events**.

1. Click **Filters**.

1. Enable showing threat events.

1. Click **Apply**.

   In the event table, events where threats were detected will appear including the new **TEST_NETCAT_LISTEN** detector.

**Detector code example**

```
package main

import (
  "context"
  "fmt"

  "github.com/gobwas/glob"
  "github.com/runtime-radar/runtime-radar/event-processor/detector/api"
  "github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

//Change metadata for a new detector
const (
  ID = "TEST_NETCAT_LISTEN"
  Name = "Use of netcat for creating incoming connections"
  Description = "The detector detects network activity related to the use of netcat to create incoming connections."
  Version = 1 Author = "CS Team"
  Contact = "email: cs@example.com"
  License = "Apache License 2.0"
)

var (
  // triggerCriteria sets Trigger Criteria as a map of event types to corresponding functions
  // to be used by the detector. If function names are not applicable to
  // a particular event type, such as "PROCESS_EXEC", leave the slice empty or use
  // the wildcard "*".
  triggerCriteria = map[string][] string {
    "PROCESS_KPROBE": {
      "inet_csk_listen_start"
    },

    // Examples:
    //
    // "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
    // In order to process all possible functions, leave the right-hand part empty or use the wildcard "*":
    // "PROCESS_EXEC": {},
    // same as:
    // "PROCESS_EXEC": {"*"},
  }
)

//Change the section with global variables for the nc template
var (
  nc = glob.MustCompile("*/nc")
)

// main is required for TinyGo to compile to Wasm
func main() {
  api.RegisterDetector(Detector {})
}

type Detector struct {}

func(d Detector) Info(ctx context.Context, req * api.InfoReq)( * api.InfoResp, error) {
  return &api.InfoResp {
    Id: ID,
    Name: Name,
    Description: Description,
    Version: Version,
    Author: Author,
    Contact: Contact,
    License: License,
  }, nil
}

func(d Detector) Detect(ctx context.Context, req * api.DetectReq)( * api.DetectResp, error) {
  // Detector information is added to DetectResp because the former is always included into a response
  // This is done to avoid an additional Wasm call on detect
  resp: = & api.DetectResp {
    Id: ID,
    Name: Name,
    Description: Description,
    Version: Version,
    Author: Author,
    Contact: Contact,

    // A default zero-value response means that nothing was detected. (This is redundant and put here just for reference
    // because Severity == api.DetectResp_NONE == 0 when omitted.)
    Severity: api.DetectResp_NONE,
  }

    event: = req.GetEvent().GetEvent()

  // Depending on the event type, the corresponding action is performed
  switch ev: = event.(type) {
    case *tetragon.GetEventsResponse_ProcessExec:
      // Nothing here
    case *tetragon.GetEventsResponse_ProcessExit:
      // Nothing here
    case *tetragon.GetEventsResponse_ProcessKprobe:
      kprobe: = ev.ProcessKprobe
      binary: = kprobe.GetProcess().GetBinary()

      function: = kprobe.GetFunctionName()
        // Extract event data

      /*New nc template to check the target function inet_csk_listen_start*/
      if !(nc.Match(binary) && function == "inet_csk_listen_start") {
        return resp, nil
      }
      // Check that netcat listens to port 8888
      args: = kprobe.GetArgs()
      if len(args) < 1 {
        return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
      }

      sport: = args[0].GetSockArg().GetSport()
      if sport == 8888 {
        resp.Severity = api.DetectResp_CRITICAL
        return resp, nil
      }

      // Nothing here
    case *tetragon.GetEventsResponse_ProcessTracepoint:
      // Nothing here
  }

    return resp, nil
}

func(d Detector) TriggerCriteria(ctx context.Context, req * api.TriggerCriteriaReq)( * api.TriggerCriteriaResp, error) {
  resp: = & api.TriggerCriteriaResp {
    Criteria: make(map[string] * api.TriggerCriteriaResp_FuncNames, len(triggerCriteria)),
  }

    for k,
  v: = range triggerCriteria {
    resp.Criteria[k] = & api.TriggerCriteriaResp_FuncNames {
      FuncNames: v
    }
  }

    return resp, nil
}
```
