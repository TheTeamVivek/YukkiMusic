from abc import ABC, abstractmethod

class Base(ABC):
  
    @abstractmethod
    async def info(self, *args, **kwargs):
        pass

    @abstractmethod
    async def download(self, *args, **kwargs):
        pass
      
    @abstractmethod
    async def valid(self, *args, **kwargs):
        pass
