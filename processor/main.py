from dataclasses import dataclass
from pydantic import BaseModel, Field

from pydantic_ai import Agent, RunContext
from pydantic_ai.models.openai import OpenAIChatModel
from pydantic_ai.providers.ollama import OllamaProvider
import typing
import enum

import json


BASE_URL = "http://localhost:11434/v1"
SYSTEM_PROMPT = """
You are an expert email triage system. Analyze the provided email and return ONLY a valid JSON object matching the TriageOutput schema.

CRITICAL: Return ONLY the JSON object. Do NOT include any markdown formatting, explanations, or additional text.

Routing policy:
- DELETE: only if the message is malicious (phishing/unsafe) OR bounced/empty noise with zero user value.
- MARK_SPAM: unwanted bulk marketing or unsolicited ads from shady or irrelevant senders, but not overtly malicious.
- MOVE: legitimate messages that belong in a user folder. Suggest a SPECIFIC folder_path using {Top/Second/OptionalThird} style.
  Smart folder_path suggestions based on sender and content:
  - Newsletters/{Brand}/{Type} (e.g., "Newsletters/Substack/Cooking", "Newsletters/LinkedIn/Updates")
  - Finance/{Institution} (e.g., "Finance/Bank", "Finance/PayPal")
  - Receipts/{Vendor} (e.g., "Receipts/Amazon", "Receipts/Netflix")
  - Social/{Platform} (e.g., "Social/LinkedIn", "Social/Facebook")
  - Work/{Company} or Work/{Type} (e.g., "Work/Job-Alerts", "Work/Company-Updates")
  - Personal/{Type} (e.g., "Personal/Family", "Personal/Friends")
  - Travel/{Service} (e.g., "Travel/Airlines", "Travel/Hotels")
  - Shopping/{Category} (e.g., "Shopping/Deals", "Shopping/Orders")
- ARCHIVE: informational but not worth a folder and not spam (e.g., one-off notifications).

Classification rules:
- Category: choose one or more from: Work, Finance, Personal, Promotions, Support, Notifications, Social, Updates, Urgent Request, Other
- Priority: High for urgent action/security alerts/deadlines; Medium for important but not urgent; Low for informational/promotional.
- Safety: Phishing/Unsafe if credential harvest/malware/suspicious domains; Spam if generic mass marketing irrelevant to the user; Safe otherwise.

Use the sender_category hint to inform your decisions:
- newsletter -> Usually MOVE to Newsletters/{Brand}
- social -> Usually MOVE to Social/{Platform}  
- finance -> Usually MOVE to Finance/{Institution}
- business -> Evaluate based on content
- personal -> Usually MOVE to Personal/ or ARCHIVE

Required JSON format (return EXACTLY this structure):
{
  "category": ["Promotions"],
  "priority": "Low",
  "safety": "Safe",
  "reasoning": "Brief explanation in 1-2 sentences",
  "organize": {
    "action": "MOVE",
    "folder_path": "Newsletters/Brand/Weekly",
    "rationale": "Brief rationale in 1-2 sentences"
  }
}

IMPORTANT: 
- ALL fields are required
- folder_path is required when action is "MOVE"
- Return ONLY the JSON object, no markdown, no explanations
"""

Category = typing.Literal[
    "Work",
    "Finance",
    "Personal",
    "Promotions",
    "Support",
    "Notifications",
    "Social",
    "Updates",
    "Urgent Request",
    "Other",
]


class ModelRegistry(enum.Enum):
    LLAMA = "llama3.1"
    DEEPSEEK = "deepseek-r1:14b"


class OrgDecision(BaseModel):
    action: typing.Literal["MOVE", "ARCHIVE", "DELETE", "MARK_SPAM"]
    folder_path: typing.Optional[str] = None  # required if action == "move"
    rationale: str = Field(..., max_length=400)


class TriageOutput(BaseModel):
    category: typing.List[Category | str]
    priority: typing.Literal["High", "Medium", "Low"]
    safety: typing.Literal["Safe", "Spam", "Phishing/Unsafe"]
    reasoning: str = Field(..., max_length=600)
    organize: OrgDecision


ollama_model = OpenAIChatModel(
    model_name=ModelRegistry.LLAMA.value,
    provider=OllamaProvider(base_url=BASE_URL),
)
agent = Agent(
    ollama_model, instructions=SYSTEM_PROMPT, output_type=TriageOutput, retries=10
)


# ---------- Dependency for Raw Payload ----------
@dataclass
class Deps:
    payload: typing.Dict[str, typing.Any]


@agent.system_prompt
def add_payload_context(ctx: RunContext[Deps]) -> str:
    """Add the email payload to the system prompt"""
    email = ctx.deps.payload

    
    return f"""
    Email to analyze:
    From: {email.get('from', 'Unknown')}
    Subject: {email.get('subject', 'No Subject')}
    Body: {email.get('bodyPlainText', email.get('bodyHtml', 'No content'))}
    Snippet: {email.get('snippet', 'No snippet')}
    """



def run_single_email_test():
    """Load the first email from emails.json and run a single triage request"""
    # Load emails from file
    with open('emails.json', 'r') as f:
        emails = json.load(f)
    
    if not emails:
        print("No emails found in emails.json")
        return
    
    # Get the first email
    first_email = emails[0]
    
    print("Processing email:")
    print(f"From: {first_email.get('from', 'Unknown')}")
    print(f"Subject: {first_email.get('subject', 'No Subject')}")
    print("-" * 50)
    
    # Create dependencies with the email payload
    deps = Deps(payload=first_email)
    
    # Run the agent
    try:
        result = agent.run_sync("Analyze this email for triage", deps=deps)
        
        # Access the output data
        data = result.output
        
        print("Triage Result:")
        print(f"Category: {data.category}")
        print(f"Priority: {data.priority}")
        print(f"Safety: {data.safety}")
        print(f"Reasoning: {data.reasoning}")
        print(f"Action: {data.organize.action}")
        if data.organize.folder_path:
            print(f"Folder: {data.organize.folder_path}")
        print(f"Rationale: {data.organize.rationale}")
        
    except Exception as e:
        print(f"Error running triage: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    run_single_email_test()

