FROM postgres
RUN ls /docker-entrypoint-initdb.d/
COPY . /docker-entrypoint-initdb.d/.
RUN rm -rf /docker-entrypoint-initdb.d/create-multiple-databases.sh
