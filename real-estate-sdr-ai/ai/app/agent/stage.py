from enum import Enum

class ConversationStage(str, Enum):
    NEW = "new"
    QUALIFYING = "qualifying"
    PROPERTY_MATCH = "property_match"
    RECOMMENDING = "recommending"
    SCHEDULING = "scheduling"
    SCHEDULED = "scheduled"
    CONVERTED = "converted"
    LOST = "lost"