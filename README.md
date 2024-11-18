# Signature Service

## Description

Signature service is a system designed to securely manage signature devices and keys and sign transaction data.

## Key Features

* Manage signature devices and keys
* Sign transaction data

## Technologies Used

* Go v1.23

## Components

### 1. Device management

* Creates new signature devices using the a device identifier, a user selected signing algorithm, and an optional label for the device.

* Stores device information in an in-memory data store. Private keys are encrypted before storage.

* Lists all signature devices.

* Retrieve a signature device by it's unique identifier.

### 2. Signature generation

* Generates a signature for the data to be signed using the keys and algorithm of the provided device identifier.

* Verifies the current signature count and the last signature generated from the signature request data.

## Setup Guide

* Clone this repository

* Install Go v1.22+

* Run `go mod tidy` to install dependencies

* Create a .env file at the root directory and set the env variables following the example in the sample.env file which can be found at the root directory.

* Run `make server` to start the application

## Run Tests

* Run `make test` to run the unit test suite. This test suite is also run in GitHub Actions CI pipeline. See .github/workflows/test.yaml.

* Run `make load_test` to run a custom load test that simulates concurrent requests for device creation and transaction signing. A report is generate and stored in the reports directory at the root of the project. This tests are not run in the CI pipeline.

## API Documentation

The API documentation for the service is available at this [Postman link](https://api.postman.com/collections/6576731-e66af949-b294-4de9-b802-c6d7fa35de2d?access_key=PMAT-01JCYYNEZQSFX8H3DS78FEEQNV)


