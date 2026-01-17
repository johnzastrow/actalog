 ---
  ## Scenario-Based Clarification Questions

  ### Scenario 1: Feature Toggle Scope

  Situation: Admin wants to test scheduling at one gym before rolling out to others.

  Question 1: Should the feature toggle be:
  - A) Global only - One switch enables/disables for all gyms
  - B) Per-gym - Each gym can independently enable/disable
  - C) Both - Global master switch + per-gym overrides

  ---
  Scenario 2: Recurrence Complexity

  Situation: A gym wants "CrossFit WOD" at 6am, 9am, noon, and 5pm every weekday, but only 9am on Saturday.

  Question 2: What recurrence patterns must be supported at launch?
  - A) Simple weekly - Single day/time per template (create 6 templates for the above)
  - B) Multi-slot weekly - Multiple days and times per template
  - C) Full RRULE - RFC 5545 subset including exceptions and complex patterns

  ---
  ### Scenario 3: Credit Charging Timing

  Situation: User reserves a spot Monday for Wednesday's class. They cancel Tuesday evening.

  Question 3: When should credits be charged?
  - A) On reservation - Credit consumed immediately; refund on cancel
  - B) On check-in - Credit consumed only when they show up
  - C) Configurable per gym - Admin chooses the policy

  If (A) or (C), what's the cancellation refund policy?
  - Full refund if canceled 24+ hours before?
  - Partial refund within 24 hours?
  - No refund?

  ---
  ### Scenario 4: Waitlist Promotion

  Situation: Class is full with 3 people waitlisted. Someone cancels at 5pm for a 6am class.

  Question 4a: How long does the promoted user have to confirm?
  - A) 10 minutes
  - B) 30 minutes
  - C) Until class starts
  - D) Configurable per gym

  Question 4b: If the promoted user doesn't confirm in time:
  - **A) Auto-promote next waitlisted user
  - **B) Spot becomes open for anyone (first-come)
  - **C) Spot stays empty until original user confirms or next cancellation

  Question 4c: How is the promoted user notified?
  - **A) Email only
  - **B) Push notification only
  - **C) Both email and push
  - **D) SMS (requires additional integration)

  ---
  ### Scenario 5: Walk-ins

  Situation: Someone shows up to class without a reservation. There's one open spot.

  Question 5: Should walk-ins be supported?
  - A) No - Must have reservation to attend
  - B) Yes, coach can add - Coach creates reservation + attendance on the spot
  - C) Yes, attendance only - Coach records attendance without reservation (affects reporting)

  ---
  ### Scenario 6: No-Show Handling

  Situation: User reserved and confirmed but didn't show up to class.

  Question 6a: When is no-show status applied?
  - **A) Automatically when session marked completed
  - **B) Coach manually marks no-show
  - **C) Automatically X minutes after class starts if not checked in

  Question 6b: What happens to credits for no-shows?
  - **A) Credit is consumed (no refund)
  - **B) Credit is restored
  - **C) Configurable per gym

  Question 6c: Should there be consequences for repeat no-shows?
  - A) No - Just track in reports
  - **B) Warning after N no-shows
  - **C) Auto-restrict booking after N no-shows in X days
  - **D) Configurable policy per gym

  ---
  ### Scenario 7: Coach Assignments

  Situation: Gym has 3 coaches. "Morning CrossFit" needs coverage Mon-Fri but coaches rotate.

  Question 7a: Can a session have multiple coaches?
  - A) No - One coach per session
  - B) Yes - Multiple coaches, one marked primary

  Question 7b: When sessions are materialized from template, how are coaches assigned?
  - A) No coach - Admin assigns after materialization
  - B) Default coach from template - Can be overridden per session
  - C) Coach schedule - Separate coach availability system (more complex)

  ---
  ### Scenario 8: Athlete Self-Check-in

  Situation: Small gym, coach is busy warming up, athletes arrive.

  Question 8: Can athletes check themselves in?
  - A) No - Only coach/admin can check in
  - B) Yes, always - Athletes can self-check-in
  - C) Configurable - Gym setting to allow/disallow self-check-in
  - D) QR code - Athlete scans code at gym to check in

  ---
  ### Scenario 9: Document Requirements

  Situation: User hasn't signed the updated waiver (new version uploaded last week).

  Question 9a: When are document requirements checked?
  - A) On reservation - Can't reserve without signed docs
  - B) On confirmation - Can reserve, must sign before confirming
  - C) On check-in - Can reserve/confirm, must sign at door

  Question 9b: When a document is updated (new version):
  - **A) Existing signatures remain valid
  - **B) Users must re-sign new version
  - **C) Admin chooses per document update

  ---
  ### Scenario 10: Class Cancellation

  Situation: Coach is sick, 6am class needs to be cancelled. 15 people are confirmed.

  Question 10a: What happens to reservations when a session is cancelled?
  - **A) All automatically cancelled, credits restored
  - **B) All automatically cancelled, no credit restore
  - **C) Admin chooses credit restore at cancellation time

  Question 10b: Should the system support rescheduling (move to different time)?
  - A) No - Cancel and create new session
  - B) Yes - Reschedule option that keeps reservations, notifies users

  ---
  ### Scenario 11: Subscription Constraints

  Situation: Gym offers "Morning Only" subscription (classes before noon) and "Unlimited" subscription.

  Question 11: Should class templates restrict which subscriptions can book?
  - A) No - Any active subscription can book any class
  - B) Yes - Templates can specify allowed subscription types
  - C) Yes, with time rules - Subscription defines allowed time windows

  ---
  ### Scenario 12: Credit Expiration

  Situation: User bought 10-credit punch card 6 months ago, used 3 credits, subscription is "indefinite."

  Question 12: Can credits expire independently of subscription?
  - A) No - Credits valid as long as subscription is active
  - B) Yes - Credits expire after X days from purchase
  - **C) Configurable per subscription type

  ---
  Please answer these questions so I can finalize the specification. You can answer with the letter choices, or provide alternative approaches I haven't considered.
