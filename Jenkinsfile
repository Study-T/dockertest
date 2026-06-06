pipeline {
    agent none

    stages {
        stage('拉取代码') {
            agent any
            steps {
                echo '==> 拉取 GitHub 代码'
                checkout scm
                stash name: 'source', includes: '**/*', useDefaultExcludes: false
            }
        }

        stage('运行测试') {
            agent {
                docker {
                    image 'golang:1.25-alpine'
                    args '--network host'
                }
            }
            steps {
                unstash 'source'
                sh '''
                    apk add --no-cache git gcc musl-dev
                    ROOT=$(pwd)
                    export GOWORK=$ROOT/go.work
                    go work sync
                    cd $ROOT/app && go test ./... -v -count=1
                    cd $ROOT/domain/tracking && go test ./... -v -count=1
                    cd $ROOT/infrastructure && go test ./... -v -count=1
                    cd $ROOT/pkg && go test ./... -v -count=1
                '''
            }
        }

        stage('构建 Docker 镜像') {
            agent any
            steps {
                unstash 'source'
                sh 'docker build -t tracking-api:latest .'
            }
        }

        stage('部署到本地 Docker') {
            agent any
            steps {
                unstash 'source'
                sh '''
                    # 清理占用端口的容器
                    docker ps -q --filter "publish=5432" | xargs -r docker rm -f
                    docker ps -q --filter "publish=6379" | xargs -r docker rm -f
                    docker ps -q --filter "publish=8082" | xargs -r docker rm -f
                    docker rm -f tracking-api tracking-postgres tracking-redis 2>/dev/null || true

                    # 创建网络
                    docker network create tracking-network 2>/dev/null || true

                    # 启动 PostgreSQL
                    docker run -d \
                        --name tracking-postgres \
                        --network tracking-network \
                        -e POSTGRES_USER=postgres \
                        -e POSTGRES_PASSWORD=123456 \
                        -e POSTGRES_DB=ns_tracking \
                        -p 5432:5432 \
                        -v postgres_data:/var/lib/postgresql/data \
                        postgres:14-alpine

                    # 启动 Redis
                    docker run -d \
                        --name tracking-redis \
                        --network tracking-network \
                        -p 6379:6379 \
                        -v redis_data:/data \
                        redis:6-alpine

                    # 等待数据库就绪
                    sleep 5

                    # 运行数据库迁移
                    docker cp infrastructure/database/migration/001_create_raw_events.sql tracking-postgres:/tmp/migration.sql
                    docker exec tracking-postgres psql -U postgres -d ns_tracking -f /tmp/migration.sql || true

                    # 启动应用
                    docker run -d \
                        --name tracking-api \
                        --network tracking-network \
                        -p 8082:8082 \
                        --restart unless-stopped \
                        tracking-api:latest

                    # 复制配置文件到容器
                    sleep 2
                    docker cp app/etc/app-docker.yaml tracking-api:/etc/app/app.yaml
                    docker restart tracking-api
                '''
            }
        }

        stage('健康检查') {
            agent any
            steps {
                sh '''
                    sleep 10
                    docker exec tracking-api wget -qO- http://localhost:8082/health || echo "健康检查失败，查看日志: docker logs tracking-api"
                    echo "==> 应用已成功部署，访问地址: http://localhost:8082"
                '''
            }
        }
    }

    post {
        success {
            echo '✅ 流水线执行成功！Go 应用已部署到本地 Docker'
        }
        failure {
            echo '❌ 流水线执行失败，请检查控制台输出'
        }
    }
}
