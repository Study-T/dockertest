pipeline {
    agent any

    environment {
        APP_NAME    = 'dockertest'
        IMAGE_LATEST = 'dockertest:latest'
        APP_PORT    = '9090'
    }

    stages {
        stage('拉取代码') {
            steps {
                echo '==> 拉取 GitHub 代码'
                checkout scm
            }
        }

        stage('安装依赖 + 运行测试') {
            steps {
                sh '''
                    pip install --no-cache-dir -r requirements.txt
                    pytest test_app.py -v
                '''
            }
        }

        stage('构建 Docker 镜像') {
            steps {
                sh "docker build -t  ."
            }
        }

        stage('部署到本地 Docker') {
            steps {
                sh '''
                    docker stop dockertest || true
                    docker rm dockertest || true
                    docker run -d \
                        --name dockertest \
                        -p 9090:8000 \
                        --restart unless-stopped \
                        dockertest:latest
                '''
            }
        }

        stage('健康检查') {
            steps {
                sh '''
                    sleep 3
                    curl -sf http://localhost:9090/health || exit 1
                    echo ""
                    echo "==> 应用已成功部署，访问地址: http://localhost:9090"
                '''
            }
        }
    }

    post {
        success {
            echo '✅ 流水线执行成功！应用已部署到本地 Docker'
        }
        failure {
            echo '❌ 流水线执行失败，请检查控制台输出'
        }
    }
}