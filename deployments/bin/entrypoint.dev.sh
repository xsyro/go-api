#!/bin/sh

while ! nc -z db 5432; do sleep 1; done;

echo 'Running migrations...'
/migrate up > /dev/null 2>&1 &

echo 'Start application...'
/app
