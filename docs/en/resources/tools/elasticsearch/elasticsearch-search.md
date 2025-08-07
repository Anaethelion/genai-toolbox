---
title: "elasticsearch-search"
type: docs
weight: 1
description: >
  Search for documents in an index.
---

# elasticsearch-search

Search for documents in an index.

This tool allows you to perform a basic search against your Elasticsearch
indices. You can specify the index to search in and the number of documents
to return.

See the [official documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/search-your-data.html) for more information.

## Example

```yaml
tools:
  search_my_index:
    kind: elasticsearch-search
    source: elasticsearch-source
    description: Use this tool to search for documents in an index.
    parameters:
      - name: index
        type: string
        description: The index to search in.
      - name: count
        type: integer
        description: The number of documents to return.
        default: 10
    query: |
      {
        "query": {
          "match_all": {}
        },
        "size": $count
      }
```

## Parameters

| **name** | **type** | **required** | **description**                                                                          |
|----------|:--------:|:------------:|------------------------------------------------------------------------------------------|
| indices  | []string |    false     | The indices to get the mapping for.                                                      |
| timeout | integer |    false     | The timeout for the query in seconds. Default is 60 (1 minute).                          |
| parameters | [parameters](../#specifying-parameters) |    false     | List of [parameters](../#specifying-parameters) that will be used with the search query. |

