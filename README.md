# operator-sdk-reachability

This project helps in generating a report of all RedHat operators using OLM to assess if they use Operator SDK or not based on the labels present in the CSV.

Operator images from the following 4 indexes have been used:
1. registry.redhat.io/redhat/redhat-marketplace-index:v4.6
2. quay.io/openshift-community-operators/catalog:latest
3. registry.redhat.io/redhat/certified-operator-index:v4.6
4. registry.redhat.io/redhat/redhat-operator-index:v4.6

The report is in the form of an Excel sheet present inside `report/` folder. It contains the name of the operators, the time at which they were created, do they have SDK labels in the CSV and other relevant data if the labels are present.

The `opm` binary utilized for the project is also present in the root of this project. It has minor modifications compared to the one existing in the operator-registry repository.

To generate a similar report at any point of time, run:

```Go
go run main.go
```

This would pull the indexes mentioned above, extract the images, unpack them and download the manifests inside `temp` folder. Further, the CSV is parsed and the SDK annotations are extracted, if present. 