# Configuration Based Workflow Library
This repo includes a library to create and execute a temporal workflow by using a json wrokflow definition with Amazon State Language syntax. Full syntax specification is located [here](https://docs.aws.amazon.com/step-functions/latest/dg/concepts-amazon-states-language.html).

# Scope
This repo contains sample code to read workflow configuration definition, and convert to execution. ASL library support following ASL syntax:
1. Pass
2. Task
3. Choice
4. Wait
5. Succeed
6. Fail
7. Parallel
8. Map

# Code Layout
### cmd/

The demo code defines its own activities, register them and use client to execute workflow defined defined in configuartion file.

### pkg/

The package folder contains state definition and state transition logic under fsm directory. Util function locate under utils directory handles parameter parsing for ASL. 