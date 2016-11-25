ECSS3COPY
==============

[![wercker status](https://app.wercker.com/status/e7bba584c295e3025cf12af4ef38302c/s/master "wercker status")](https://app.wercker.com/project/byKey/e7bba584c295e3025cf12af4ef38302c)

OVERVIEW
--------------

ECSS3COPY is a tool developped in Golang to copy objects from one bucket to another one using S3 copy operations.

Metadata search queries can also be indicated to select the objects to copy.

RUN
--------------

To start the application, run:
```
docker run -it djannot/ecss3copy ./ecss3copy --help
Usage:
  ecss3copy [OPTIONS]

Application Options:
  -e, --endpoint= The ECS endpoint
  -u, --user=     The ECS object user
  -p, --password= The ECS object user password
  -s, --source=   The ECS source bucket
  -t, --target=   The ECS target bucket
  -m, --maxkeys=  The number of keys to retrieve simultaneously from the ECS
                  source bucket (default: 100)
  -q, --query=    The ECS metadata search query to select the objects from the
                  source bucket
  -v, --verbose   Verbose mode also display the object successfully copies

Help Options:
  -h, --help      Show this help message
```

LICENSING
--------------

Licensed under the Apache License, Version 2.0 (the “License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
