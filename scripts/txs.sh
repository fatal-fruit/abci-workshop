#!/bin/bash

cosmappd tx bank send beatrice $(cosmappd keys show alice -a) 100uatom -y --output json
#cosmappd tx bank send beatrice $(cosmappd keys show bob -a) 10uatom -y --output json
cosmappd tx bank send bob $(cosmappd keys show alice -a) 100uatom -y --output json
cosmappd tx bank send alice $(cosmappd keys show beatrice -a) 100uatom -y --output json
