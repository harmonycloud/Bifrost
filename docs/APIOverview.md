# API Overview

## Namespace IP pool

### **Create**

Create a namespace IP pool.

#### HTTP Request

`POST /nsippool`

#### Body Parameters

| Parameter | Description                                                  | Required                         |
| --------- | ------------------------------------------------------------ | -------------------------------- |
| name      | Name of the IP pool to be created. Name must be unique within a namespace. | True                             |
| namespace | Namespace defines the space within each name must be unique. | True                             |
| subnet    | IP addresses in the IP pool belongs to this subnet.          | True                             |
| start     | The first IP address of the IP pool.                         | True (or use cidr instead)       |
| end       | The last IP address of the IP pool.                          | True (or use cidr instead)       |
| cidr      | Another representation of IP addresses in the IP pool.       | True (or use start,end instead ) |
| gateway   | Gateway of all pods using IP addresses from this IP pool.    | True                             |
| vlanid    | VLAN ID of the IP pool. (Integer)                            | True                             |

#### Request Body Example

```json
{
    "name": "example-ns-pool",
    "namespace": "default",
    "subnet": "192.168.1.0/24",
    "start": "192.168.1.5",
    "end": "192.168.1.24",
    "gateway": "192.168.1.1",
    "vlanid": 100
}
```

#### Response

| Result  | Description                                                  | Code |
| ------- | ------------------------------------------------------------ | ---- |
| Success | Successfully create the IP pool, return its' full information. | 200  |
| Existed | IP pool's name isn't unique within a namespace.              | 400  |
| Overlap | Some IPs are already in another namespace IP pool.           | 400  |
| Error   | Error occurred when creating IP pool.                        | 400  |

#### Successful Response Example

```json
{
    "name": "example-ns-pool",
    "namespace": "default",
    "subnet": "192.168.1.0/24",
    "start": "192.168.1.5",
    "end": "192.168.1.24",
    "gateway": "192.168.1.1",
    "vlanid": 100,
    "total": 20,
    "PodMap": null
}
```



### **Read**

Read information of a namespace IP pool.

#### HTTP Request

`GET /nsippool/{namespace}/{name}`

#### Path Parameters

| Parameter | Description                                                 |
| --------- | ----------------------------------------------------------- |
| namespace | Object name and auth scope, such as for teams and projects. |
| name      | Name of the namespace IP pool.                              |

#### Response

| Result  | Description                           | Code |
| ------- | ------------------------------------- | ---- |
| Success | Get information of requested IP pool. | 200  |
| Failed  | Can not find requested IP pool.       | 400  |



### **List all**

Read information of all namespace IP pools in the given namespace.

#### HTTP Request

`GET /getAllPool/{namespace}`

#### Path Parameters

| Parameter | Description                                                 |
| --------- | ----------------------------------------------------------- |
| namespace | Object name and auth scope, such as for teams and projects. |

#### Response

| Result  | Description                                             | Code |
| ------- | ------------------------------------------------------- | ---- |
| Success | Get information of all IP pools in the given namespace. | 200  |
| Failed  | Can not find IP pools in the given namespace.           | 400  |



### **Delete**

Delete information of a namespace IP pool.

#### HTTP Request

`DELETE /nsippool/{namespace}/{name}`

#### Path Parameters

| Parameter | Description                                                 |
| --------- | ----------------------------------------------------------- |
| namespace | Object name and auth scope, such as for teams and projects. |
| name      | Name of the namespace IP pool to be deleted.                |

#### Response

| Result  | Description                                | Code |
| ------- | ------------------------------------------ | ---- |
| Success | Successfully delete the namespace IP pool. | 200  |
| Failed  | Error occurred when deleting the IP pool.  | 400  |



## Service IP pool

### **Create**

Create a service IP pool, usually is a part of a namespace IP pool.

#### HTTP Request

`POST /serviceIPPool`

#### Body Parameters

| Parameter    | Description                                                  | Required                         |
| ------------ | ------------------------------------------------------------ | -------------------------------- |
| name         | Name of the IP pool to be created. Name must be unique within a namespace. | True                             |
| namespace    | Namespace defines the space within each name must be unique. | True                             |
| start        | The first IP address of the IP pool.                         | True (or use cidr instead)       |
| end          | The last IP address of the IP pool.                          | True (or use cidr instead)       |
| cidr         | Another representation of IP addresses in the IP pool.       | True (or use start,end instead ) |
| serviceName  | Name of the kubernetes service using the IP pool.            | True                             |
| nsIPPoolName | Name of the namespace IP pool that the service IP pool belongs to. | True                             |

#### Request Body Example

```json
{
    "name": "example-svc-pool",
    "namespace": "default",
    "start": "192.168.1.5",
    "end": "192.168.1.14",
    "serviceName": "example-service",
    "nsIPPoolName": "example-ns-pool"
}
```

#### Response

| Result   | Description                                                  | Code |
| -------- | ------------------------------------------------------------ | ---- |
| Success  | Successfully create the IP pool, return its' full information. | 200  |
| Existed  | IP pool's name isn't unique within a namespace.              | 400  |
| Overlap  | Some IPs are already in another namespace IP pool.           | 400  |
| NotFound | Can not find requested namespace IP pool.                    | 400  |
| Error    | Error occurred when creating IP pool.                        | 400  |

#### Successful Response Example

```json
{
    "name": "example-svc-pool",
    "namespace": "default",
    "start": "192.168.1.5",
    "end": "192.168.1.14",
    "serviceName": "example-service",
    "nsIPPoolName": "example-ns-pool",
    "total": 10
}
```



### **Read**

Read information of a service IP pool.

#### HTTP Request

`GET /serviceIPPool/{namespace}/{name}`

#### Path Parameters

| Parameter | Description                                                 |
| --------- | ----------------------------------------------------------- |
| namespace | Object name and auth scope, such as for teams and projects. |
| name      | Name of the service IP pool.                                |

#### Response

| Result  | Description                           | Code |
| ------- | ------------------------------------- | ---- |
| Success | Get information of requested IP pool. | 200  |
| Failed  | Can not find requested IP pool.       | 400  |



### **Delete**

Delete information of a service IP pool.

#### HTTP Request

`DELETE /serviceIPPool/{namespace}/{name}`

#### Path Parameters

| Parameter | Description                                                 |
| --------- | ----------------------------------------------------------- |
| namespace | Object name and auth scope, such as for teams and projects. |
| name      | Name of the service IP pool to be deleted.                  |

#### Response

| Result  | Description                                                  | Code |
| ------- | ------------------------------------------------------------ | ---- |
| Success | Successfully delete the service IP pool.                     | 200  |
| Failed  | Error occurred when deleting the IP pool or there are pods still running with IPs in the service IP pool. | 400  |

