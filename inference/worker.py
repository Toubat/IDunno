import grpc
import sys
import argparse
from concurrent import futures

from api_pb2_grpc import add_InferenceServiceServicer_to_server, WorkerServiceStub
from inference import InferenceServiceServer

def main():
    global WORKER_PORT

    parser = argparse.ArgumentParser()
    parser.add_argument("--port", type=int, default=6000)
    parser.add_argument("--filepath", type=str, default="images/")
    args = parser.parse_args()

    # max receive size = 100MB
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10), options=[
        ('grpc.max_receive_message_length', 100 * 1024 * 1024),
        ('grpc.max_send_message_length', 100 * 1024 * 1024),
    ])
    add_InferenceServiceServicer_to_server(InferenceServiceServer(), server)
    server.add_insecure_port("[::]:%d" % args.port)
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    main()