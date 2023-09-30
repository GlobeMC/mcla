#!/bin/sh

git add . && \
git commit -m 'dev' && \
git tag -d dev && \
git tag dev && \
git push -f origin dev
