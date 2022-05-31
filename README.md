# golang-crawlstackoverflow
- RESTful Web API CRUD using GORM in Golang
- Crawl data with Golang Colly
- Crawl Scheduler with jasonlvhit/gocron
- Docker
- Deploy in Kubernetes
    + kubectl create -f mysql-secret.yaml
    + kubectl apply -f mysql-db-pv.yaml
    + kubectl apply -f mysql-db-pvc.yaml
    + kubectl apply -f mysql-db-deployment.yaml
    + kubectl apply -f mysql-db-service.yaml
    + kubectl apply -f app-mysql-deployment.yaml
    + kubectl apply -f app-mysql-service.yaml
    + kubectl get pods
    + kubectl get services
    + minikube start
    + minikube dashboard
    + kubectl port-forward service/fullstack-app-mysql 28015:8080 => 127.0.0.1:28015
