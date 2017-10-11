#!/bin/bash
echo "1:STDOUT"
>&2 echo "2:STDERR"
echo "3:STDOUT"
>&2 echo "4:STDERR"
>&2 echo "5:STDERR"
echo "6:STDOUT"
