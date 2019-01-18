#!/bin/bash

if [ "$TRAVIS_PULL_REQUEST" != "false" ]; then
  echo "Pull request detected - running unit tests"
  ip addr show
  make test
elif [ "$TRAVIS_BRANCH" == "develop" ] || [ "$TRAVIS_BRANCH" == "master" ]; then
  # This is when a merge happens - ideally we'd split up develop and master but we only have a single production server so no need 
  echo "Deploying services to production"
else 
  echo "No deployment necessary"
fi
