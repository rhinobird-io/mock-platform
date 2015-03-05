FROM ruby:2.1

run mkdir -p /tmp/app
ADD ./repo/Gemfile /tmp/app/

WORKDIR /tmp/app

run bundle install

WORKDIR /

run rm -rf /tmp/repo

