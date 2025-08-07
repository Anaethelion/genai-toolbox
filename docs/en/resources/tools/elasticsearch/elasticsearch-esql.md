---
title: "elasticsearch-esql"
type: docs
weight: 2
description: >
  Execute ES|QL queries.
---

# elasticsearch-esql

Execute ES|QL queries.

This tool allows you to execute ES|QL queries against your Elasticsearch
cluster. You can use this to perform complex searches and aggregations.

See the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/esql-query-api.html) for more information.

## Example

```yaml
tools:
  query_my_index:
    kind: elasticsearch-esql
    source: elasticsearch-source
    description: Use this tool to execute ESQL queries.
    query: |
      FROM $index
      | KEEP *
      | SORT $field $sort_order
    parameters:
      - name: index
        type: string
        description: The index to query.
        required: true
      - name: field
        type: string
        description: The field to sort by.
        required: true
      - name: sort_order
        type: string
        description: The order to sort the results.
        required: true
```

## Parameters

| **name**   | **type** | **required** | **description**                                                                          |
|------------|:--------:|:------------:|------------------------------------------------------------------------------------------|
| query      |  string  |     true     | The field to sort by.                                                                    |
| format | string | false | The format of the query. Default is json. Valid values are csv, json, tsv, txt, yaml, cbor, smile, or arrow.          |
| timeout | integer |    false     | The timeout for the query in seconds. Default is 60 (1 minute).                          |
| parameters | [parameters](../#specifying-parameters) |    false     | List of [parameters](../#specifying-parameters) that will be used with the ES\|QL query. |

