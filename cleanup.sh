kubectl delete deployments gif-maker --context dev
kubectl delete deployments gif-maker --context khightower
kubectl delete deployments gif-maker --context production
git tag -d 0.0.1
git push --delete origin 0.0.1
