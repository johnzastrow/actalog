# Scheduling Classes

The system should now allow scheduling of classes, managing reservations, check-ins, and punch-card type memberships.

In addition to user types of admin and athlete, we will introduce a new user role: coach.



#### 1. As an athlete I should be able to:
1.1 see a list of classes that I can reserve in the future.
1.1.1 the listing will show the information shown in image_007.png
1.1.2 click a card from the list shows the details of the class shown in image_006.png
1.2 see list of classes I have completed
1.3 reserve classes for me to take in the future
1.4 cancel classes in the future that they have reserved
1.5 if canceled, and athlete has a "punch card" membership, the credit should be returned to the athlete's membership
1.6 An athlete is a user who is not a coach (new user role) or an admin
1.7 athletes may not cancel a reservation from the past and the class will reduce the credits if the membership is a punch card type. 

#### 2. As a coach I should be able to
2.1 Create standard workouts
2.2 Check atheletes into workouts - even if they have not previously reserved


#### 10. Memberships
10.1 Athlete punch card memberships have an expiration date that defaults to one year from when they are created
10.2 Admins can change membership expiration dates
10.3 Admins can add or remove credits from punch card memberships
10.4 All actions against memberships is capture in the audit log
10.5 When an athlete reserves a class, a credit is deducted from their punch card membership
10.6 When an athlete cancels a class reservation, a credit is returned to their punch card membership
10.7 If an athlete does not show up for a class they reserved, the credit is not returned to their punch card membership
10.7 Athletes with active punch card memberships may only reserve classes if they have at least one credit remaining
10.8 Athletes can track their remaining credits and expiration date from their profile page



#### 11. Roles
1.6 An athlete is a user who is not a coach (new user role) or an admin
