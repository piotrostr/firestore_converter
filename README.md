# firestore_converter

## Environment Vars

```
FIREBASE_SERVICE_ACCOUNT=/path/to/service-account.json
FIREBASE_DATABASE_ID=some-staging-database
```

## Limitations

I would not recommend to run this on production databases, this is a script for
snapshotting the production DB into the staging environment

In highly nested collections structurethe dump/load methods might need to be
adjusted, this tool dumps subcollections into root level

```sh
go build .
```

dump with

```sh
FIREBASE_DATABASE_ID="(default)" firestore_converter --dump
```

and then

```sh
firestore_converter --load
```
