#!/bin/sh

echo -e "\n2. 应用配置文件..."
sudo kubectl apply -f namespace.yaml
sudo kubectl apply -f deployment.yaml
sudo kubectl apply -f service.yaml

# 重启部署
sudo kubectl rollout restart deploy trade-itg -n pay-ns

echo -e "\n3. 等待部署完成..."
sudo kubectl rollout status deployment trade-itg -n pay-ns --timeout=120s

echo -e "\n=== 部署完成 ==="
echo -e "\nPod 状态："
sudo kubectl get pods -n pay-ns -o wide

echo -e "\nService 状态："
sudo kubectl get svc -n pay-ns

echo -e "\n最近日志："
sudo kubectl logs -n pay-ns -l app=trade-itg --tail=20
