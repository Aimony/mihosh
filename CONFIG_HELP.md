# mihosh 配置说明

## 配置文件位置

- Windows: `C:\Users\你的用户名\.mihosh\config.yaml`
- Linux / macOS: `~/.mihosh/config.yaml`

## 可用配置项

```yaml
api_address: http://127.0.0.1:9090
secret: your-secret-here
test_url: http://www.gstatic.com/generate_204
timeout: 5000
proxy_address: http://127.0.0.1:7890
```

## CLI 设置命令

```bash
mihomo config set api-address http://127.0.0.1:9090
mihomo config set secret your-secret-here
mihomo config set test-url http://www.gstatic.com/generate_204
mihomo config set timeout 5000
mihomo config set proxy-address http://127.0.0.1:7890
```

## 初始化

```bash
mihomo config init
```
