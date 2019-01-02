# Sir            

[![CircleCI](https://circleci.com/gh/komuw/sir.svg?style=svg)](https://circleci.com/gh/komuw/sir)
[![codecov](https://codecov.io/gh/komuw/sir/branch/master/graph/badge.svg)](https://codecov.io/gh/komuw/sir)
[![GoDoc](https://godoc.org/github.com/komuw/sir?status.svg)](https://godoc.org/github.com/komuw/sir)
[![Go Report Card](https://goreportcard.com/badge/github.com/komuw/Sir)](https://goreportcard.com/report/github.com/komuw/sir)          


Sir, is a TCP reverse proxy that checks for regressions in your services/apps.               
It's name is derived from the late(and great) Kenyan hip hop artiste, [E-Sir](https://en.wikipedia.org/wiki/E-Sir). 


It's currently work in progress, API will remain unstable for sometime.        


#### premise
Sir finds potential bugs in your service/s using running instances of your new code and your old code side by side.          
Sir behaves as a proxy and multicasts whatever requests it receives to each of the running instances.             
It then compares:     
1. the requests, and reports any regressions that may surface from those comparisons.       
2. the responses, and reports any regressions that may surface from those comparisons.        

The premise for Sir is that:    
1. If two implementations of the service send “similar” requests for a sufficiently large and diverse set of responses, then the two implementations can be treated as equivalent and the newer implementation is regression-free.       
2. If two implementations of the service return “similar” responses for a sufficiently large and diverse set of requests, then the two implementations can be treated as equivalent and the newer implementation is regression-free.        

```sh

                request                        forward-request
| Client |   ---------------->     | Sir|  ---------------------->    | Your App |
| Client |                         | Sir|                             | Your App |
                                                                 (Your app processes request)
| Client |                         | Sir|       response              | Your App |
| Client |                         | Sir|  <----------------------    | Your App |
| Client |                         | Sir|                             | Your App |
                          (Sir analyzes requests/responses)
| Client |                         | Sir|                             | Your App |
| Client |                         | Sir|                             | Your App |
| Client |   forward-response      | Sir|                             | Your App |
| Client |   <----------------     | Sir|                             | Your App |
| Client |                         | Sir|                             | Your App |
| Client |                         | Sir|                             | Your App |
                        (Sir reports any reqressions found)
```      



#### prior art
1. https://github.com/twitter/diffy      


