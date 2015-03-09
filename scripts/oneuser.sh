#!/bin/bash

sudo docker exec -it tw-platform sh -c "irb << END
require './app'
User.first.email
END"
