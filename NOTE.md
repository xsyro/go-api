### Running docker-compose in development mode

When you use a Docker volume with the line `- .:/src in your docker-compose.yml` file, you're effectively mounting your local directory (where your `docker-compose.yml` file is located) to `/src` inside the container. This means that the container will use the `entrypoint.sh` script from your local host when it runs, because that local file will overwrite the one inside the image.

However, it's important to note that even though the container will use the `entrypoint.sh` script from your local machine, the permissions (in this case, the executable permission) that you set on the file inside the Docker image with `chmod` won't carry over to the file on your local machine. The permissions on your local file remain the same.

Running the `RUN chmod +x ./deployments/bin/entrypoint.sh` command sets the executable permission on the `entrypoint.sh` script inside the Docker image. But when you start the container with `docker-compose up`, the `entrypoint.sh` script from your local host will overwrite the one inside the container (because of the volume mount).

If the `entrypoint.sh` script on your local machine isn't already executable, you might run into the "permission denied" error, because the `chmod` changes in the `Dockerfile` don't affect your local file.

To avoid this, you should ensure the `entrypoint.sh` script is executable on your local machine by running:

```bash
chmod +x ./deployments/bin/entry.sh
```
on your local host before starting the container with `docker-compose up`.