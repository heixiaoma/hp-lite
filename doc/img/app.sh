#!/bin/bash

# 配置部分
APP_NAME="hp-server-2.0-SNAPSHOT.jar"           # 应用名称
APP_JAR="./hp-server-2.0-SNAPSHOT.jar"  # Java应用JAR包路径
JAVA_OPTS="-XX:MaxDirectMemorySize=256m" # JVM参数
PID_FILE="/var/run/${APP_NAME}.pid"  # 存储PID的文件路径

# 获取当前PID
get_pid() {
    if [ -f "${PID_FILE}" ]; then
        cat "${PID_FILE}"
    else
        echo "PID file not found."
        return 1
    fi
}

# 启动应用
start_app() {
    echo "Starting ${APP_NAME}..."
    nohup java ${JAVA_OPTS} -jar "${APP_JAR}" > ./app.log 2>&1 &
    echo $! > "${PID_FILE}"
    echo "${APP_NAME} started with PID $(get_pid)."
}

# 停止应用
stop_app() {
    local pid=$(get_pid)
    if [ -z "${pid}" ]; then
        echo "No PID found, ${APP_NAME} might not be running."
        return 1
    fi

    echo "Stopping ${APP_NAME} with PID ${pid}..."
    kill "${pid}"
    sleep 5
    if ps -p "${pid}" > /dev/null; then
        echo "Process ${pid} is still running. Forcing termination..."
        kill -9 "${pid}"
    fi
    rm -f "${PID_FILE}"
    echo "${APP_NAME} stopped."
}

# 主要逻辑
if [ "$1" == "restart" ]; then
    stop_app
    start_app
elif [ "$1" == "start" ]; then
    start_app
elif [ "$1" == "stop" ]; then
    stop_app
else
    echo "Usage: $0 {start|stop|restart}"
    exit 1
fi
