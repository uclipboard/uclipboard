import asyncio
import websockets
import json
import time
import sys


async def send_messages(uri):
    async with websockets.connect(uri) as websocket:
        for i in range(300):
            try:
                message = {
                    "type": "push",
                    "content": f"hello world from {i}",
                    "hostname": "hostname",
                    "content_type": "text"
                }
                await websocket.send(json.dumps(message))
                print(f"sent {i+1}", end='\r')
                await asyncio.sleep(0.1)
            except Exception as e:
                print(f"send failed: {str(e)}")
                break

if __name__ == "__main__":
    print("begin to send websocket message.")
    start_time = time.time()
    asyncio.run(send_messages(sys.argv[1]))
    end_time = time.time()
    print(f"\ndone! duration: {end_time - start_time:.2f}s")
