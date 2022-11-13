from locust import HttpUser, task

class TestRun(HttpUser):
    @task
    def Test_Run(self):
        self.client.get("/hi")
