FROM fnproject/ruby:dev

WORKDIR /function
ADD Gemfile* /function/
RUN bundle install

ADD . /function/

ENTRYPOINT ["ruby", "func.rb"]
