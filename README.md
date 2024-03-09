# Toy Dapr based cicd stuff

I want to see if I can write ci cd in a normal programming language
and connect it to github webhooks (or, as A FALLBACK, github actions)

The toy might be github-reliant, but the end goal is to use a standard of some sort.

## Design

Gitops enabler. Runs scripts on PR open, push the branch that optionally has a PR open, on merge to main etc.

### Core / Engine

The main engine will be running a Go program the user wrote. That Go program will import this package, and pass a struct, implementing an interface we define.

example:

```go
package main

import (
    actions "github.com/drornir/dapr-actions"
)

func main () {
    myServer := actions.NewServerFromHandler(/*...*/)

    err := actions.Serve(&myServer)
    if err != nil {
        myServer.Logger().Error(err)
        os.Exit(1)
    }
}
```

Ideally, in your company, you will have a custom wrapper for `.NewServerFromHandler(/*...*/)` which will include things for secrets, plugins etc.

### Orchestrator / Runtime

Something needs to run those go scripts, which is what I want to use dapr for. Every git repo needs to be monitored (polling or webhooks). On event coming from git, the system needs to decide what to execute, and execute that.

#### Architechture

```
event-streamer --> Orchestrator -=> Actions --> Process/Run
                    ^^^^                
                     actionsctl
```

On the left side, we have an **event streamer**, which is an HTTP webhook listener which pipes the http payload into an in-cluster event system/service.

to the right of that, the **Orchestrator** maps the event into a set of actions on a set of repositories. Usually, one action on a single repo.

Actions are defined in scripts like in the Core section above. They are checked into source control, meaning they are associated to that git repo (by default). This metadata is held and managed by the Orchstrator.

Each action is traslated into invoking a process (in the broad sense of the word - from a goroutine, container, dapr actor to whatever).

Each process is monitored for its success or fail state, saving logs and all the nice stuff we have in gh actions GUI.

An **actionsctl** should be written in order to query the system and give high level commands to the system. This is like kubectl and many other k8s native apps. I want it to be in the same standard format that the events come in, so the Orchestrator doesn't need to speak multiple languages. The orchestrator already needs to speak "K8s", so maybe passing CRD like structures is best.

