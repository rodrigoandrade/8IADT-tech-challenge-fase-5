from enum import Enum


class AgentType(str, Enum):
    QUALIFICATION = "qualification"
    KNOWLEDGE = "knowledge"
    PROPERTY = "property"